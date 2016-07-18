# Go up until we are in /zanaduu3 directory
function cdroot()
{
	while [[ $PWD != '/' && ${PWD##*/} != 'zanaduu3' ]]; do cd ..; done
}
cdroot

FILE_PATH=src/site/static
sass ${FILE_PATH}/scss/arbital.scss ${FILE_PATH}/css/compiled/arbital.css
