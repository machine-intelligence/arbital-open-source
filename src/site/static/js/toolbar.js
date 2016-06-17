'use strict';

// toolbar directive displays the toolbar at the top of each page
app.directive('arbToolbar', function($mdSidenav, $http, $mdPanel, $location, $compile, $rootScope, $timeout,
		$q, $mdMedia, arb) {
	return {
		templateUrl: versionUrl('static/html/toolbar.html'),
		scope: {
			loadingBarValue: '=',
			currentUrl: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.isTinyScreen = !$mdMedia('gt-xs');
			$scope.selectedUpdatesButton = -1;

			$scope.doAutofocus = function() {
				return !arb.isTouchDevice && !arb.urlService.hasLoadedFirstPage;
			};

			// Called when a search result is selected
			$scope.searchResultSelected = function(result) {
				if (result) {
					arb.urlService.goToUrl(arb.urlService.getPageUrl(result.pageId));
				}
			};

			$scope.getSignupUrl = function() {
				return '/signup/?continueUrl=' + encodeURIComponent($location.absUrl());
			};

			$scope.showSignupButton = function() {
				return !arb.userService.userIsLoggedIn() && $location.path().indexOf('/signup/') != 0;
			};

			// Open RHS menu
			$scope.toggleRightMenu = function() {
				$mdSidenav('right').toggle();
			};

			$scope.logout = function() {
				Cookies.remove('masteryMap');
				Cookies.remove('arbital');
				window.location.reload();
			};

			// Hide toolbar in the edit screen
			$scope.$on('$locationChangeSuccess', function() {
				$scope.hide = $location.path().indexOf('/edit') === 0;
			});
			$scope.hide = $location.path().indexOf('/edit') === 0;

			$scope.showNotifications = function(ev) {
				arb.userService.user.newNotificationCount = 0;
				showPanel(
					ev,
					'/notifications/',
					'.notifications-icon',
					'<arb-updates-panel post-url="/json/notifications/" hide-title="true" num-to-display="20" more-link="/notifications"></arb-udpates-panel>'
				);
				if (!arb.isTouchDevice) {
					$scope.selectedUpdatesButton = 0;
				}
			};

			$scope.showAchievements = function(ev) {
				arb.userService.user.newAchievementCount = 0;
				showPanel(
					ev,
					'/achievements/',
					'.achievements-icon',
					'<arb-hedons-mode-panel hide-title="true" num-to-display="20"></arb-hedons-mode-panel>'
				);
				if (!arb.isTouchDevice) {
					$scope.selectedUpdatesButton = 1;
				}
			};

			$scope.showMaintenanceUpdates = function(ev) {
				arb.userService.user.maintenanceUpdateCount = 0;
				showPanel(
					ev,
					'/maintain/',
					'.maintenance-updates-icon',
					'<arb-updates-panel post-url="/json/maintain/" hide-title="true" num-to-display="20" more-link="/maintain"></arb-updates-panel>'
				);
				if (!arb.isTouchDevice) {
					$scope.selectedUpdatesButton = 2;
				}
			};

			var showPanel = function(ev, fullPageUrl, relPosElement, panelTemplate) {
				if (!$mdMedia('gt-sm')) {
					arb.urlService.goToUrl(fullPageUrl);
					return;
				}

				var position = $mdPanel.newPanelPosition()
					.relativeTo(relPosElement)
					.addPanelPosition($mdPanel.xPosition.ALIGN_END, $mdPanel.yPosition.BELOW);
				var config = {
					template: panelTemplate,
					position: position,
					panelClass: 'popover-panel md-whiteframe-8dp',
					openFrom: ev,
					clickOutsideToClose: true,
					escapeToClose: true,
					focusOnOpen: false,
					zIndex: 200000,
					onRemoving: function() {
						$scope.selectedUpdatesButton = -1;
					},
				};
				var panel = $mdPanel.create(config);
				panel.open();

				$scope.$on('$locationChangeSuccess', function() {
					panel.close();
				});
			};
		},
	};
});
