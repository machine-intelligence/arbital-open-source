# Go up until we are in /zanaduu3 directory
function cdroot()
{
	while [[ $PWD != '/' && ${PWD##*/} != 'zanaduu3' ]]; do cd ..; done
}
cdroot

FILE_PATH=src/site/static/js
jscs --fix --config ${FILE_PATH}/.jscsrc ${FILE_PATH}
