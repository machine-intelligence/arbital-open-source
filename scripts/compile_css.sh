# Go up until we are in /default directory
function cdroot()
{
	while [[ $PWD != '/' && ${PWD##*/} != 'default' ]]; do cd ..; done
}
cdroot

FILE_PATH=src/site/static
lessc ${FILE_PATH}/less/base.less > ${FILE_PATH}/css/base.css
lessc ${FILE_PATH}/less/claim.less > ${FILE_PATH}/css/claim.css
lessc ${FILE_PATH}/less/claims.less > ${FILE_PATH}/css/claims.css
