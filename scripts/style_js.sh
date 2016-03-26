# Go up until we are in /default directory
function cdroot()
{
	while [[ $PWD != '/' && ${PWD##*/} != 'default' ]]; do cd ..; done
}
cdroot

FILE_PATH=src/site/static/js
jscs --fix --config ${FILE_PATH}/.jscsrc ${FILE_PATH}
