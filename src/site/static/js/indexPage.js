'use strict';

// arb-index directive displays a set of featured domains
app.directive('arbIndex', function($http, arb) {
	return {
		templateUrl: 'static/html/indexPage.html',
		controller: function($scope) {
			$scope.arb = arb;
			$scope.readTab = 0;
			$scope.writeTab = 0;

			$scope.selectReadTab = function(tab) {
				$scope.readTab = tab;
			};

			$scope.selectWriteTab = function(tab) {
				$scope.writeTab = tab;
			};

			// Find out if we show the continueWriting panel
			arb.stateService.postData('/json/continueWriting/', {},
				function(data) {
					$scope.showContinueWritingPanel = data.result.modeRows.length > 0;
				});
		},
	};
});
