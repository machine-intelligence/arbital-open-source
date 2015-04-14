/* angular.tmpl.js is a .tmpl file that is inserted as a <script> into the
	<header> portion of html pages that use angular. It defines the zanaduu module
	and ZanaduuCtrl, which are used on every page. */
{{define "angular"}}
<script>

var app = angular.module("zanaduu", ["ngResource", "ui.bootstrap", "RecursionHelper"]);
app.config(function($interpolateProvider){
	$interpolateProvider.startSymbol("{[{").endSymbol("}]}");
});

// Children factory gets children pages for the given parent(s).
app.factory("Children", ["$resource", function($resource){
	return $resource("/json/children/", {}, {
		get: {method:"GET"},
	});
}]);

// ZanaduuCtrl is used across all pages.
// It stores all the loaded pages.
// It provides multiple helper functions for working with pages.
app.controller("ZanaduuCtrl", ["$scope", function ($scope) {
	// All loaded pages.
	$scope.pageMap = {};
	// Pages loaded along with html & js.
	$scope.initialPages = {
		{{range $k,$v := .PageMap}}
			"{{$k}}": {{GetPageJson $v}},
		{{end}}
	};
	console.log($scope.initialPages);
	// Get the url corresponding to the given page.
	$scope.getPageUrl = function(page) {
		return "/pages/" + page.Alias;
	};
}]);

// likesPageTitle displays likes span followed by page's title span.
app.directive("likesPageTitle", [function() {
	return {
		templateUrl: "/static/html/likesPageTitle.html",
	};
}]);


</script>
{{end}}
