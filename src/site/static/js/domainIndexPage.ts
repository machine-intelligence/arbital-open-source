'use strict';

import app from './angular.ts';
import {arraysSortFn} from './util.ts';

// arb-index directive displays the main page
app.directive('arbDomainIndex', function($http, $mdMedia, arb) {
	return {
		templateUrl: versionUrl('static/html/domainIndexPage.html'),
		scope: {
			domain: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;

			// Tab stuff
			$scope.readTab = 0;
			$scope.writeTab = 0;
			$scope.selectReadTab = function(tab) {
				$scope.readTab = tab;
			};
			$scope.selectWriteTab = function(tab) {
				$scope.writeTab = tab;
			};
		},
	};
});
