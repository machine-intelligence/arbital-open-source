'use strict';

// toolbar directive displays the toolbar at the top of each page
app.directive('arbToolbar', function($mdSidenav, $http, $mdPanel, $location, $compile, $rootScope, $timeout,
		$q, $mdMedia, arb) {
	return {
		templateUrl: 'static/html/toolbar.html',
		scope: {
			loadingBarValue: '=',
			currentUrl: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.isTinyScreen = !$mdMedia('gt-xs');

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

			$scope.showAchievements = function(ev) {
				showPanel(
					ev,
					'/achievements/',
					'.achievements-icon',
					'<arb-hedons-mode-panel hide-title="true" num-to-display="100"></arb-hedons-mode-panel>'
				);
			};

			$scope.showMaintenanceUpdates = function(ev) {
				showPanel(
					ev,
					'/maintain/',
					'.maintenance-updates-icon',
					'<arb-maintenance-mode-panel hide-title="true" num-to-display="100"></arb-maintenance-mode-panel>'
				);
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
					panelClass: 'popover-panel',
					openFrom: ev,
					clickOutsideToClose: true,
					escapeToClose: true,
					focusOnOpen: false,
					zIndex: 200000,
				};
				$mdPanel.open(config);
			}
		},
	};
});
