# Go up until we are in /default directory
function cdroot()
{
	while [[ $PWD != '/' && ${PWD##*/} != 'default' ]]; do cd ..; done
}
cdroot

FILE_PATH=src/site/static
sass ${FILE_PATH}/scss/arbital.scss ${FILE_PATH}/css/compiled/arbital.css
