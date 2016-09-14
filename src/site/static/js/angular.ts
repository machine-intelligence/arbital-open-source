'use strict';

// Set up angular module.
var app = angular.module('arbital', ['ngMaterial', 'ngResource',
		'ngMessages', 'ngSanitize', 'RecursionHelper', 'as.sortable']);
export default app;

app.config(function($locationProvider, $mdIconProvider, $mdThemingProvider) {
	// Convert "rgb(#,#,#)" color to "#hex"
	var rgb2hex = function(rgb) {
		if (rgb === undefined)
			return '#000000';
		rgb = rgb.match(/^rgb\((\d+),\s*(\d+),\s*(\d+)\)$/);
		function hex(x) {
			return ('0' + parseInt(x).toString(16)).slice(-2);
		}
		return '#' + hex(rgb[1]) + hex(rgb[2]) + hex(rgb[3]);
	};
	// Create themes, by getting the colors from our css files
	$mdThemingProvider.definePalette('arb-primary-theme', $mdThemingProvider.extendPalette('teal', {
		'500': rgb2hex($('#primary-color').css('border-top-color')),
		'300': rgb2hex($('#primary-color').css('border-right-color')),
		'800': rgb2hex($('#primary-color').css('border-bottom-color')),
		'A100': rgb2hex($('#primary-color').css('border-left-color')),
		'contrastDefaultColor': 'light',
		'contrastDarkColors': ['300'],
	}));
	$mdThemingProvider.definePalette('arb-accent-theme', $mdThemingProvider.extendPalette('deep-orange', {
		'A200': rgb2hex($('#accent-color').css('border-top-color')),
		'A100': rgb2hex($('#accent-color').css('border-right-color')),
		'A400': rgb2hex($('#accent-color').css('border-bottom-color')),
		'A700': rgb2hex($('#accent-color').css('border-left-color')),
		'contrastDefaultColor': 'dark',
		'contrastLightColors': [],
	}));
	$mdThemingProvider.definePalette('arb-warn-theme', $mdThemingProvider.extendPalette('red', {
		'500': rgb2hex($('#warn-color').css('border-top-color')),
		'300': rgb2hex($('#warn-color').css('border-right-color')),
		'800': rgb2hex($('#warn-color').css('border-bottom-color')),
		'A100': rgb2hex($('#warn-color').css('border-left-color')),
		'contrastDefaultColor': 'light',
		'contrastDarkColors': ['300'],
	}));
	// Set the theme
	$mdThemingProvider.theme('default')
	.primaryPalette('arb-primary-theme', {
		'default': '500',
		'hue-1': '300',
		'hue-2': '800',
		'hue-3': 'A100',
	})
	.accentPalette('arb-accent-theme', {
		'default': 'A200',
		'hue-1': 'A100',
		'hue-2': 'A400',
		'hue-3': 'A700',
	})
	.warnPalette('arb-warn-theme', {
		'default': '500',
		'hue-1': '300',
		'hue-2': '800',
		'hue-3': 'A100',
	});

	// Set up custom icons
	$mdIconProvider.icon('arbital_logo', 'static/icons/arbital-logo.svg', 40)
		.icon('comment_plus_outline', 'static/icons/comment-plus-outline.svg')
		.icon('comment_question_outline', 'static/icons/comment-question-outline.svg')
		.icon('facebook_box', 'static/icons/facebook-box.svg')
		.icon('file_outline', 'static/icons/file-outline.svg')
		.icon('format_header_pound', 'static/icons/format-header-pound.svg')
		.icon('cursor_pointer', 'static/icons/cursor-pointer.svg')
		.icon('link_variant', 'static/icons/link-variant.svg')
		.icon('slack', 'static/icons/slack.svg')
		.icon('thumb_up_outline', 'static/icons/thumb-up-outline.svg')
		.icon('thumb_down_outline', 'static/icons/thumb-down-outline.svg');

	$locationProvider.html5Mode(true);
});

app.run(function($http, $location, arb) {
	// Set up mapping from URL path to specific controllers
	arb.urlService.addUrlHandler('/', {
		name: 'IndexPage',
		handler: function(args, $scope) {
			if ($scope.subdomain) {
				// Get the private domain index page data
				$http({method: 'POST', url: '/json/domainPage/', data: JSON.stringify({})})
				.success($scope.getSuccessFunc(function(data) {
					$scope.indexPageIdsMap = data.result;
					return {
						title: arb.stateService.pageMap[$scope.subdomain].title + ' - Private Domain',
						content: $scope.newElement('<arb-group-index group-id=\'' + data.result.domainId +
							'\' ids-map=\'::indexPageIdsMap\'></arb-group-index>'),
					};
				}))
				.error($scope.getErrorFunc('domainPage'));
			} else {
				// Get the index page data
				$http({method: 'POST', url: '/json/index/'})
				.success($scope.getSuccessFunc(function(data) {
					return {
						title: '',
						content: $scope.newElement('<arb-index></arb-index>'),
					};
				}))
				.error($scope.getErrorFunc('index'));
			}
		},
	});
	arb.urlService.addUrlHandler('/project/', {
		name: 'ProjectPage',
		handler: function(args, $scope) {
			$http({method: 'POST', url: '/json/index/'})
			.success($scope.getSuccessFunc(function(data) {
				return {
					title: '',
					content: $scope.newElement('<arb-project></arb-project>'),
				};
			}))
			.error($scope.getErrorFunc('index'));
		},
	});
	arb.urlService.addUrlHandler('/achievements/', {
		name: 'AchievementsPage',
		handler: function(args, $scope) {
			$http({method: 'POST', url: '/json/default/'})
			.success($scope.getSuccessFunc(function(data) {
				return {
					title: 'Achievements',
					content: $scope.newElement('<arb-hedons-mode-page></arb-hedons-mode-page>'),
				};
			}))
			.error($scope.getErrorFunc('default'));
		},
	});
	arb.urlService.addUrlHandler('/adminDashboard/', {
		name: 'AdminDashboardPage',
		handler: function(args, $scope) {
			var postData = {};
			// Get the data
			$http({method: 'POST', url: '/json/adminDashboardPage/', data: JSON.stringify(postData)})
			.success($scope.getSuccessFunc(function(data) {
				$scope.adminDashboardData = data.result;
				return {
					title: 'Admin dashboard',
					content: $scope.newElement('<arb-admin-dashboard-page data=\'::adminDashboardData\'></arb-admin-dashboard-page>'),
				};
			}))
			.error($scope.getErrorFunc('adminDashboardPage'));
		},
	});
	arb.urlService.addUrlHandler('/dashboard/', {
		name: 'DashboardPage',
		handler: function(args, $scope) {
			var postData = {};
			// Get the data
			$http({method: 'POST', url: '/json/dashboardPage/', data: JSON.stringify(postData)})
			.success($scope.getSuccessFunc(function(data) {
				$scope.data = data.result;
				return {
					title: 'Your dashboard',
					content: $scope.newElement('<arb-dashboard-page data=\'::data\'></arb-dashboard-page>'),
				};
			}))
			.error($scope.getErrorFunc('dashboardPage'));
		},
	});
	arb.urlService.addUrlHandler('/discussion/', {
		name: 'DiscussionPage',
		handler: function(args, $scope) {
			$http({method: 'POST', url: '/json/default/'})
			.success($scope.getSuccessFunc(function(data) {
				return {
					title: 'Discussions',
					content: $scope.newElement('<arb-discussion-mode-page></arb-discussion-mode-page>'),
				};
			}))
			.error($scope.getErrorFunc('default'));
		},
	});
	arb.urlService.addUrlHandler('/edit/:alias?/:alias2?', {
		name: 'EditPage',
		handler: function(args, $scope) {
			var pageAlias = args.alias;
			// Check if we are already editing this page.
			// TODO: this is kind of hacky. We need a better solution (I hope React will help us with this).
			var primaryPage = arb.stateService.primaryPage;
			if (primaryPage && primaryPage.pageId in arb.stateService.editMap &&
					(primaryPage.pageId == pageAlias || primaryPage.alias == pageAlias)) {
				return true;
			}

			// Load the last edit for the pageAlias.
			var loadEdit = function() {
				var quickParentId = $location.search().parentId;
				arb.pageService.loadEdit({
					pageAlias: pageAlias,
					additionalPageIds: quickParentId ? [quickParentId] : undefined,
					specificEdit: $location.search().edit ? +$location.search().edit : 0,
					success: $scope.getSuccessFunc(function(): any {
						// Find the page in the editMap (have to search through it manually
						// because we don't index pages by alias in editmap)
						var page;
						for (var pageId in arb.stateService.editMap) {
							page = arb.stateService.editMap[pageId];
							if (page.alias == pageAlias || page.pageId == pageAlias) {
								break;
							}
						}
						if ($location.search().alias) {
							// Set page's alias
							page.alias = $location.search().alias;
							$location.replace().search('alias', undefined);
						}

						// If the page has a pending edit, and we are not on it, reload
						if (page.proposalEditNum > 0 && page.proposalEditNum != page.edit) {
							$location.replace().search('edit', page.proposalEditNum);
							window.location.href = $location.url();
							return {};
						}

						arb.urlService.ensureCanonPath(arb.urlService.getEditPageUrl(page.pageId));
						arb.stateService.primaryPage = page;

						// Called when the user is done editing the page.
						$scope.doneFn = function(result) {
							var page = arb.stateService.editMap[result.pageId];
							// if the page is (now / still) live, go (to / back to) it
							if ((page.wasPublished && !result.deletedPage) || !result.discard) {
								arb.urlService.goToUrl(arb.urlService.getPageUrl(page.pageId, {
									useEditMap: true,
									markId: $location.search().markId,
								}));

								if (result.pageId in arb.stateService.pageMap &&
										!arb.stateService.pageMap[result.pageId].isSubscribedAsMaintainer) {
									arb.popupService.showToast({
										text: 'Maintain this page?',
										scope: $scope,
										normalButton: {
											text: 'Subscribe',
											icon: 'build',
											callbackText: 'subscribeAsMaintainer()',
										},
									});
								}
							} else {
								// if the page is deleted or unpublished, go home
								// TODO: ideally we should navigate to whatever page we came from before opening the editor
								$location.path('/');
							}
						};

						$scope.subscribeAsMaintainer = function() {
							$http({method: 'POST', url: '/updateSubscription/', data: JSON.stringify({
								toId: page.pageId,
								isSubscribed: true,
								asMaintainer: true,
							})});
							arb.stateService.pageMap[page.pageId].isSubscribed = true;
							arb.stateService.pageMap[page.pageId].isSubscribedAsMaintainer = true;

							arb.popupService.showToast({
								text: 'Subscribed as maintainer',
								scope: $scope,
							});
						};

						return {
							removeBodyFix: true,
							title: 'Edit ' + (page.title ? page.title : 'New Page'),
							content: $scope.newElement('<arb-edit-page class=\'full-height\' page-id=\'' + page.pageId +
								'\' done-fn=\'doneFn(result)\' layout=\'column\'></arb-edit-page>'),
						};
					}),
					error: $scope.getErrorFunc('edit'),
				});
			};

			// Load a new page.
			var getNewPage = function() {
				var type = $location.search().type;
				$location.replace().search('type', undefined);
				// Create a new page to edit
				arb.pageService.getNewPage({
					type: type,
					success: function(newPageId) {
						arb.urlService.goToUrl(arb.urlService.getEditPageUrl(newPageId, {
							parentId: $location.search().parentId,
						}), {replace: true});
					},
					error: $scope.getErrorFunc('newPage'),
				});
			};

			// Need to call /default/ in case we are creating a new page
			// TODO(alexei): have /newPage/ return /default/ data along with /edit/ data
			$http({method: 'POST', url: '/json/default/'})
			.success($scope.getSuccessFunc(function(data): any {
				// Redirect non-logged in users to sign up
				if (!arb.userService.userIsLoggedIn()) {
					arb.urlService.goToUrl('/signup?continueUrl=' + encodeURIComponent($location.url()), {replace: true});
					return {};
				}

				if (pageAlias) {
					loadEdit();
				} else {
					getNewPage();
				}
				return {
					title: 'Edit Page',
				};
			}))
			.error($scope.getErrorFunc('default'));
		},
	});
	arb.urlService.addUrlHandler('/explore/:pageAlias', {
		name: 'ExplorePage',
		handler: function(args, $scope) {
			$http({method: 'POST', url: '/json/explore/', data: JSON.stringify({pageAlias: args.pageAlias})})
			.success($scope.getSuccessFunc(function(data) {
				var page = arb.stateService.pageMap[data.result.pageId];
				return {
					title: 'Explore ' + page.title,
					content: $scope.newElement('<arb-explore-page page-id=\'' + data.result.pageId +
						'\'></arb-explore-page>'),
				};
			}))
			.error($scope.getErrorFunc('explore'));
		},
	});
	arb.urlService.addUrlHandler('/groups/', {
		name: 'GroupsPage',
		handler: function(args, $scope) {
			$http({method: 'POST', url: '/json/groups/'})
			.success($scope.getSuccessFunc(function(data) {
				return {
					title: 'Groups',
					content: $scope.newElement('<arb-groups-page></arb-groups-page>'),
				};
			}))
			.error($scope.getErrorFunc('groups'));
		},
	});
	arb.urlService.addUrlHandler('/learn/:pageAlias?', {
		name: 'LearnPage',
		handler: function(args, $scope) {
			// Get the primary page data
			var postData = {
				pageAliases: [],
				onlyWanted: $location.search()['only_wanted'] === '1', // jscs:ignore requireDotNotation
			};
			var continueLearning = false;
			if (args.pageAlias) {
				postData.pageAliases.push(args.pageAlias);
			} else if ($location.search().path) {
				postData.pageAliases = postData.pageAliases.concat($location.search().path.split(','));
			} else if (arb.pathService.path) {
				postData.pageAliases = arb.pathService.path.pageIds;
				continueLearning = true;
			}

			$http({method: 'POST', url: '/json/learn/', data: JSON.stringify(postData)})
			.success($scope.getSuccessFunc(function(data) {
				var primaryPage = undefined;
				if (args.pageAlias) {
					primaryPage = arb.stateService.pageMap[args.pageAlias];
					arb.urlService.ensureCanonPath('/learn/' + primaryPage.alias);
				}

				$scope.learnPageIds = data.result.pageIds;
				$scope.learnOptionsMap = data.result.optionsMap;
				$scope.learnTutorMap = data.result.tutorMap;
				$scope.learnRequirementMap = data.result.requirementMap;
				return {
					title: 'Learn ' + (primaryPage ? primaryPage.title : ''),
					content: $scope.newElement('<arb-learn-page continue-learning=\'::' + continueLearning +
						'\' page-ids=\'::learnPageIds\'' +
						'\' options-map=\'::learnOptionsMap\'' +
						' tutor-map=\'::learnTutorMap\' requirement-map=\'::learnRequirementMap\'' +
						'></arb-learn-page>'),
				};
			}))
			.error($scope.getErrorFunc('learn'));
		},
	});
	arb.urlService.addUrlHandler('/login/', {
		name: 'LoginPage',
		handler: function(args, $scope) {
			$http({method: 'POST', url: '/json/default/'})
			.success($scope.getSuccessFunc(function(data) {
				if (arb.userService.user.id) {
					window.location.href = arb.urlService.getDomainUrl();
				}
				return {
					title: 'Log In',
					content: $scope.newElement('<div class=\'md-whiteframe-1dp capped-body-width\'><arb-login></arb-login></div>'),
				};
			}))
			.error($scope.getErrorFunc('default'));
		},
	});
	arb.urlService.addUrlHandler('/maintain/', {
		name: 'MaintainPage',
		handler: function(args, $scope) {
			$http({method: 'POST', url: '/json/default/'})
			.success($scope.getSuccessFunc(function(data) {
				return {
					title: 'Maintain',
					content: $scope.newElement('<arb-maintenance-mode-page></arb-maintenance-mode-page>'),
				};
			}))
			.error($scope.getErrorFunc('default'));
		},
	});
	arb.urlService.addUrlHandler('/newsletter/', {
		name: 'NewsletterPage',
		handler: function(args, $scope) {
			$http({method: 'POST', url: '/json/newsletter/'})
			.success($scope.getSuccessFunc(function(data) {
				return {
					title: 'Newsletter',
					content: $scope.newElement('<arb-newsletter></arb-newsletter>'),
				};
			}))
			.error($scope.getErrorFunc('newsletter'));
		},
	});
	arb.urlService.addUrlHandler('/notifications/', {
		name: 'NotificationsPage',
		handler: function(args, $scope) {
			$http({method: 'POST', url: '/json/default/'})
			.success($scope.getSuccessFunc(function(data) {
				return {
					title: 'Notifications',
					content: $scope.newElement('<arb-bell-updates-page></arb-bell-updates-page>'),
				};
			}))
			.error($scope.getErrorFunc('default'));
		},
	});
	arb.urlService.addUrlHandler('/p/:alias/:alias2?', {
		name: 'PrimaryPage',
		handler: function(args, $scope) {
			// Check if we are just switching to the primary lens
			// TODO: this is kind of hacky. We need a better solution (I hope React will help us with this).
			var primaryPage = arb.stateService.primaryPage;
			if (primaryPage && !(primaryPage.pageId in arb.stateService.editMap) &&
					(primaryPage.pageId == args.alias || primaryPage.alias == args.alias)) {
				return true;
			}

			// Get the primary page data
			var postData = {
				pageAlias: args.alias,
				lensId: $location.search().l,
				markId: $location.search().markId,
				hubId: $location.search().hubId,
				pathPageId: $location.search().pathPageId,
				// Load the path if it's not loaded already
				pathInstanceId: arb.stateService.path ? undefined : $location.search().pathId,
			};
			$http({method: 'POST', url: '/json/primaryPage/', data: JSON.stringify(postData)})
			.success($scope.getSuccessFunc(function(data): any {
				var primaryPageId = data.result.primaryPageId;
				var page = arb.stateService.pageMap[primaryPageId];
				var pageTemplate = '<arb-primary-page></arb-primary-page>';

				if (data.result.path) {
					arb.stateService.path = data.result.path;
				} else if (arb.stateService.path && !$location.search().pathId) {
					// We are off the path. Forget the path if it was finished.
					var isFinished = arb.stateService.path.isFinished ||
						arb.stateService.path.progress >= arb.stateService.path.pages.length-1;
					if (isFinished) {
						arb.stateService.path = undefined;
					}
				}

				if (!page) {
					page = arb.stateService.deletedPagesMap[postData.pageAlias];
					if (page) {
						if (page.mergedInto) {
							arb.urlService.goToUrl(arb.urlService.getPageUrl(page.mergedInto));
						} else {
							arb.urlService.goToUrl(arb.urlService.getEditPageUrl(postData.pageAlias));
						}
						return {};
					}
					return {
						title: 'Not Found',
						error: 'Page doesn\'t exist or you don\'t have permission to view it.',
					};
				}

				// For comments, redirect to the primary page
				if (page.isComment()) {
					// TODO: allow BE to catch this case and send correct data, so we don't have to reload
					arb.urlService.goToUrl(arb.urlService.getPageUrl(page.getCommentParentPage().pageId), {replace: true});
					return {};
				}

				// If this page has been merged into another, go there
				if (page.mergedInto) {
					arb.urlService.goToUrl(arb.urlService.getPageUrl(page.mergedInto));
					return {};
				}

				// If the page is a user page, get the additional data about user
				// - Recently created by me page ids.
				// - Recently created by me comment ids.
				// - Recently edited by me page ids.
				// - Top pages by me
				if (arb.userService.userMap[page.pageId]) {
					$scope.userPageIdsMap = data.result;
					pageTemplate = '<arb-user-page user-id=\'' + page.pageId +
							'\' user_page_data=\'::userPageIdsMap\'></arb-user-page>';
				}

				arb.stateService.primaryPage = page;
				arb.urlService.ensureCanonPath(arb.urlService.getPageUrl(page.pageId, {lensId: data.result.lensId}));
				arb.pathService.primaryPageChanged();
				return {
					title: page.title,
					content: $scope.newElement(pageTemplate),
				};
			}))
			.error($scope.getErrorFunc('primaryPage'));
		},
	});
	arb.urlService.addUrlHandler('/read/', {
		name: 'ReadModePage',
		handler: function(args, $scope) {
			$http({method: 'POST', url: '/json/default/'})
			.success($scope.getSuccessFunc(function(data) {
				return {
					title: 'Read',
					content: $scope.newElement('<arb-read-mode-page></arb-read-mode-page>'),
				};
			}))
			.error($scope.getErrorFunc('default'));
		},
	});
	arb.urlService.addUrlHandler('/recentChanges/', {
		name: 'RecentChangesPage',
		handler: function(args, $scope) {
			$http({method: 'POST', url: '/json/default/'})
			.success($scope.getSuccessFunc(function(data) {
				return {
					title: 'Recent changes',
					content: $scope.newElement('<arb-recent-changes-page></arb-recent-changes-page>'),
				};
			}))
			.error($scope.getErrorFunc('default'));
		},
	});
	arb.urlService.addUrlHandler('/requisites/', {
		name: 'RequisitesPage',
		handler: function(args, $scope) {
			$http({method: 'POST', url: '/json/requisites/'})
			.success($scope.getSuccessFunc(function(data) {
				return {
					title: 'Requisites',
					content: $scope.newElement('<arb-requisites-page></arb-requisites-page>'),
				};
			}))
			.error($scope.getErrorFunc('requisites'));
		},
	});
	arb.urlService.addUrlHandler('/settings/', {
		name: 'SettingsPage',
		handler: function(args, $scope) {
			$http({method: 'POST', url: '/json/settingsPage/'})
			.success($scope.getSuccessFunc(function(data) {
				if (data.result) {
					$scope.domains = data.result.domains;
					// Convert invitesSent object to array for ease in angular
					$scope.invitesSent = [];
					for (var key in data.result.invitesSent) {
						$scope.invitesSent.push(data.result.invitesSent[key]);
					}
				}
				return {
					title: 'Settings',
					content: $scope.newElement('<arb-settings-page domains="::domains" ' +
						'invites-sent="::invitesSent"></arb-settings-page>'),
				};
			}))
			.error($scope.getErrorFunc('settingsPage'));
		},
	});
	arb.urlService.addUrlHandler('/signup/', {
		name: 'SignupPage',
		handler: function(args, $scope) {
			$http({method: 'POST', url: '/json/default/'})
			.success($scope.getSuccessFunc(function(data) {
				if (arb.userService.user.id) {
					window.location.href = arb.urlService.getDomainUrl();
				}
				return {
					title: 'Sign Up',
					content: $scope.newElement('<arb-signup></arb-signup>'),
				};
			}))
			.error($scope.getErrorFunc('default'));
		},
	});
});

// simpleDateTime filter converts our typical date&time string into local time.
app.filter('simpleDateTime', function() {
	return function(input) {
		return moment.utc(input).local().format('LT, l');
	};
});
// smartDateTime converts date&time into a relative string or a date string
// depending on how long the event was
app.filter('smartDateTime', function() {
	return function(input) {
		if (moment.utc().diff(moment.utc(input), 'days') <= 7) {
			return moment.utc(input).fromNow();
		}
		if (moment.utc().diff(moment.utc(input), 'months') <= 10) {
			return moment.utc(input).local().format('MMM D');
		}
		return moment.utc(input).local().format('MMM D, YYYY');
	};
});
// relativeDateTime converts date&time into a relative string, e.g. "5 days ago"
app.filter('relativeDateTime', function() {
	return function(input) {
		return moment.utc(input).fromNow();
	};
});

// numSuffix filter converts a number string to a 2 digit number with a suffix, e.g. K, M, G
app.filter('numSuffix', function() {
	return function(input) {
		var num = +input;
		if (num >= 100000) return (Math.round(num / 100000) / 10) + 'M';
		if (num >= 100) return (Math.round(num / 100) / 10) + 'K';
		return input;
	};
});

// shorten filter shortens a string to the given number of characters
app.filter('shorten', function() {
	return function(input, charCount) {
		if (!input || input.length <= charCount) return input;
		var s = input.substring(0, charCount);
		var lastSpaceIndex = s.lastIndexOf(' ');
		if (lastSpaceIndex < 0) return s + '...';
		return input.substring(0, lastSpaceIndex) + '...';
	};
});
