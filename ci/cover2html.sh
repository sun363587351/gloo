#!/bin/bash -ex

OUTPUTDIR=artifacts/coverage


mkdir -p $OUTPUTDIR
REPORT=$OUTPUTDIR/index.html

MEGACOV=$OUTPUTDIR/megacov.megacoverprofile
echo 'mode: atomic' > "$MEGACOV"

echo "<html><body>" > $REPORT

find . -name "*.coverprofile" -print0 | while read -d $'\0' file; do
    FDIR=$(dirname "$file")
    COVDIR=$OUTPUTDIR/$FDIR
    mkdir -p $COVDIR
    FNAME=$(basename $file)
    OUTFILE="$COVDIR/$FNAME".html
    go tool cover -html="$file" -o "$OUTFILE"
     tail --lines=+2 $file >> $MEGACOV
    echo '<div><a href="'$FDIR/$FNAME'.html">'$FNAME"</a>" >> $REPORT
    echo "<pre>" >> $REPORT
    go tool cover -func="$file" >> $REPORT
    echo "</pre></div>" >> $REPORT

done

echo "</body></html>" >> $REPORT

go tool cover -html="$MEGACOV" -o $OUTPUTDIR/megacov.html