// Our style sheets
require('scss/arbital.scss');

// All AngularJS templates.
var templates = require.context(
		'html',
		true,
		/\.html$/
)
templates.keys().forEach(function(key) {
	templates(key);
});

// Base AngularJS app
require('js/angular.ts');

// AngularJS services
require('js/analyticsService.ts');
require('js/autocompleteService.ts');
require('js/diffService.ts');
require('js/markService.ts');
require('js/markdownService.ts');
require('js/masteryService.ts');
require('js/pageService.ts');
require('js/pathService.ts');
require('js/popoverService.ts');
require('js/popupService.ts');
require('js/signupService.ts');
require('js/stateService.ts');
require('js/urlService.ts');
require('js/userService.ts');
require('js/arbService.ts');

// AngularJS directives
require('js/toolbar.ts');
require('js/arbDirectives.ts');
require('js/markdown.ts');
require('js/voteBar.ts');
require('js/page.ts');
require('js/lens.ts');
require('js/pageDiscussion.ts');
require('js/subpage.ts');
require('js/pageImprovement.ts');
require('js/relationships.ts');
require('js/reqRelationships.ts');
require('js/childRelationships.ts');
require('js/masteryList.ts');
require('js/multipleChoice.ts');
require('js/checkbox.ts');
require('js/tableOfContents.ts');
require('js/queryInfo.ts');
require('js/markInfo.ts');
require('js/answers.ts');
require('js/marks.ts');
require('js/hiddenText.ts');
require('js/editDiff.ts');
require('js/exploreTreeNode.ts');
require('js/pathEditor.ts');
require('js/pathNav.ts');
require('js/learnMore.ts');
require('js/changeSpeedButton.ts');
require('js/hubPageGui.ts');
require('js/hubPageFooter.ts');

// AngularJS controllers
require('js/arbitalController.ts');
require('js/editPageDialog.ts');
require('js/feedbackDialog.ts');
require('js/rhsButtons.ts');

// Page specific directives
require('js/primaryPage.ts');
require('js/editPage.ts');
require('js/hedonsMode.ts');
require('js/writeMode.ts');
require('js/indexPage.ts');
require('js/groupsPage.ts');
require('js/userPage.ts');
require('js/dashboardPage.ts');
require('js/loginPage.ts');
require('js/signupPage.ts');
require('js/requisitesPage.ts');
require('js/learnPage.ts');
require('js/settingsPage.ts');
require('js/settingsInviteTab.ts');
require('js/adminDashboardPage.ts');
require('js/readMode.ts');
require('js/recentChanges.ts');
require('js/discussionMode.ts');
require('js/newsletterPage.ts');
require('js/explorePage.ts');
require('js/updateRows.ts');
require('js/updatesMode.ts');
