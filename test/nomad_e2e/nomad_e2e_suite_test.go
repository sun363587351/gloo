package nomad_e2e

import (
	"os"
	"testing"

	"time"

	"bytes"
	"context"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/consul/api"
	vaultapi "github.com/hashicorp/vault/api"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo/pkg/log"
	"github.com/solo-io/gloo/pkg/storage"
	"github.com/solo-io/gloo/pkg/storage/consul"
	"github.com/solo-io/gloo/pkg/storage/dependencies"
	"github.com/solo-io/gloo/pkg/storage/dependencies/vault"
	"github.com/solo-io/gloo/test/helpers"
	"github.com/solo-io/gloo/test/helpers/local"
	"github.com/solo-io/gloo/test/nomad_e2e/utils"
)

func TestConsul(t *testing.T) {
	if os.Getenv("RUN_NOMAD_TESTS") != "1" {
		log.Printf("This test downloads and runs nomad consul and vault. It is disabled by default. " +
			"To enable, set RUN_NOMAD_TESTS=1 in your env.")
		return
	}
	RegisterFailHandler(func(message string, callerSkip ...int) {
		var logs string
		for _, task := range []string{"control-plane", "ingress"} {
			l, err := utils.Logs(nomadInstance, "gloo", task)
			logs += l + "\n"
			if err != nil {
				logs += "error getting logs for " + task + ": " + err.Error()
			}
		}
		addr, err := helpers.ConsulServiceAddress("ingress", "admin")
		if err == nil {
			configDump, err := helpers.Curl(addr, helpers.HttpOpts{Path: "/config_dump"})
			if err == nil {
				logs += "\n\n\n" + configDump + "\n\n\n"
			}
		}

		log.Printf("\n****************************************" +
			"\nLOGS FROM THE BOYS: \n\n" + logs + "\n************************************")
		Fail(message, callerSkip...)
	})
	log.DefaultOut = GinkgoWriter
	RunSpecs(t, "Nomad Suite")
}

var (
	vaultFactory  *localhelpers.VaultFactory
	vaultInstance *localhelpers.VaultInstance

	consulFactory  *localhelpers.ConsulFactory
	consulInstance *localhelpers.ConsulInstance

	nomadFactory  *localhelpers.NomadFactory
	nomadInstance *localhelpers.NomadInstance

	gloo    storage.Interface
	secrets dependencies.SecretStorage

	ingressAddrHttp string

	err error
)

var _ = BeforeSuite(func() {
	vaultFactory, err = localhelpers.NewVaultFactory()
	helpers.Must(err)
	vaultInstance, err = vaultFactory.NewVaultInstance()
	helpers.Must(err)
	err = vaultInstance.Run()
	helpers.Must(err)

	consulFactory, err = localhelpers.NewConsulFactory()
	helpers.Must(err)
	consulInstance, err = consulFactory.NewConsulInstance()
	helpers.Must(err)
	consulInstance.Silence()
	err = consulInstance.Run()
	helpers.Must(err)

	nomadFactory, err = localhelpers.NewNomadFactory()
	helpers.Must(err)
	nomadInstance, err = nomadFactory.NewNomadInstance()
	helpers.Must(err)
	nomadInstance.Silence()
	err = nomadInstance.Run()
	helpers.Must(err)

	gloo, err = consul.NewStorage(api.DefaultConfig(), "gloo", time.Second)
	helpers.Must(err)

	vaultCfg := vaultapi.DefaultConfig()
	vaultCfg.Address = "http://127.0.0.1:8200"
	cli, err := vaultapi.NewClient(vaultCfg)
	helpers.Must(err)
	cli.SetToken(vaultInstance.Token())

	secrets = vault.NewSecretStorage(cli, "gloo", time.Second)

	err = utils.SetupNomadForE2eTest(nomadInstance, true)
	helpers.Must(err)
	ingressAddrHttp, err = helpers.ConsulServiceAddress("ingress", "http")
	helpers.Must(err)
})

var _ = AfterSuite(func() {
	if err := utils.TeardownNomadE2e(); err != nil {
		log.Warnf("FAILED TEARING DOWN: %v", err)
	}

	vaultInstance.Clean()
	vaultFactory.Clean()

	consulInstance.Clean()
	consulFactory.Clean()

	nomadInstance.Clean()
	nomadFactory.Clean()

})

// only meant for http, for https use curl
func expectHttpResponse(opts helpers.HttpOpts, expectedStatus int, expectedBody string, timeout time.Duration) {
	var (
		body string
	)

	Eventually(func() (int, error) {
		req, err := http.NewRequest(opts.Method, "http://"+ingressAddrHttp+opts.Path, bytes.NewBufferString(opts.Body))
		helpers.Must(err)
		ctx, cancel := context.WithTimeout(req.Context(), timeout/time.Duration(5))
		defer cancel()

		req = req.WithContext(ctx)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0, err
		}
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return 0, err
		}

		body = string(b)
		log.Debugf("got response body: %v", body)

		defer res.Body.Close()

		return res.StatusCode, nil
	}, timeout, "2s").Should(Equal(expectedStatus))

	Expect(body).To(ContainSubstring(expectedBody))
}
