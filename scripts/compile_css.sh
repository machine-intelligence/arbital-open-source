# Go up until we are in /default directory
function cdroot()
{
	while [[ $PWD != '/' && ${PWD##*/} != 'default' ]]; do cd ..; done
}
cdroot

FILE_PATH=src/go/site/static
lessc ${FILE_PATH}/less/rewards.less > ${FILE_PATH}/css/rewards.css
lessc ${FILE_PATH}/less/refer.less > ${FILE_PATH}/css/refer.css
lessc ${FILE_PATH}/less/base.less > ${FILE_PATH}/css/base.css
