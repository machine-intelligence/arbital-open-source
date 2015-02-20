# Go up until we are in /default directory
function cdroot()
{
	while [[ $PWD != '/' && ${PWD##*/} != 'default' ]]; do cd ..; done
}
cdroot

FILE_PATH=src/site/static
lessc ${FILE_PATH}/less/base.less > ${FILE_PATH}/css/compiled/base.css
lessc ${FILE_PATH}/less/page.less > ${FILE_PATH}/css/compiled/page.css
lessc ${FILE_PATH}/less/recentPages.less > ${FILE_PATH}/css/compiled/recentPages.css
lessc ${FILE_PATH}/less/pagedown.less > ${FILE_PATH}/css/compiled/pagedown.css
