#!/bin/sh
set -e

#
# This script downloads our original font files from their source repos
# and puts them in our originalMedia folder.
#

err_report() {
  echo "Error on line $(caller)" >&2
}

trap err_report ERR

ORIGINAL_FONTS_DIR="./originalMedia/fonts"

# update these if there is a new version
FONT_URLS=(
"https://github.com/googlefonts/opensans/blob/27d060e1aad6886daeda67629ee28189f795f534/fonts/variable/OpenSans%5Bwdth%2Cwght%5D.ttf?raw=true"
"https://github.com/googlefonts/opensans/blob/27d060e1aad6886daeda67629ee28189f795f534/fonts/variable/OpenSans-Italic%5Bwdth%2Cwght%5D.ttf?raw=true"
"https://github.com/andrew-paglinawan/QuicksandFamily/blob/db6de44878582966f45a0debaef10d57108d93a7/fonts/Quicksand%5Bwght%5D.ttf?raw=true"
)


echo ""
echo "###################################################"
echo "# Download font files"
echo "###################################################"
echo ""

mkdir -p $ORIGINAL_FONTS_DIR

for URL in ${FONT_URLS[@]}; do
	wget -L $URL \
		--directory-prefix=$ORIGINAL_FONTS_DIR \
		--quiet \
		--timestamping \
		--show-progress
done

echo ""
echo "###################################################"
echo "# Remove '?raw=true' filename suffix"
echo "###################################################"
echo ""

# Iterate over all files in directory with filetype ending in "?raw=true"
for file in $ORIGINAL_FONTS_DIR/*?raw=true; do
	# Remove "?raw=true" from file name and store in variable
	new_name=$(echo $file | sed 's/?raw=true//')

	# Overwrite existing file with new name
	mv -v $file $new_name
done

echo "Renaming files complete"
