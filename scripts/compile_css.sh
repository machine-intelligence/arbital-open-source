# Go up until we are in /default directory
function cdroot()
{
	while [[ $PWD != '/' && ${PWD##*/} != 'default' ]]; do cd ..; done
}
cdroot

FILE_PATH=src/site/static
lessc ${FILE_PATH}/less/base.less > ${FILE_PATH}/css/base.css
lessc ${FILE_PATH}/less/question.less > ${FILE_PATH}/css/question.css
