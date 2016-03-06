app.run(function($rootScope, urlService) {
	$rootScope.$on('$locationChangeSuccess', function (event, url) {
		urlService.resolveUrl();
	};
app.run(function($rootScope, $http, $compile, pageService, urlService) {
	urlService
	.when("/", {
		template: "",
		controller: "IndexPageController",
		handler: function (args) {
			if (urlService.subdomain) {
				// Get the private domain index page data
				$http({method: "POST", url: "/json/domainPage/", data: JSON.stringify({})})
				.success(urlService.getSuccessFunc(function(data){
					$rootScope.indexPageIdsMap = data.result;
					return {
						title: pageService.pageMap[urlService.subdomain].title + " - Private Domain",
						element: $compile("<arb-group-index group-id='" + data.result.domainId +
							"' ids-map='::indexPageIdsMap'></arb-group-index>")($rootScope),
					};
				}))
				.error(urlService.getErrorFunc("domainPage"));
			} else {
				// Get the index page data
				$http({method: "POST", url: "/json/index/"})
				.success(urlService.getSuccessFunc(function(data){
					urlService.featuredDomains = data.result.featuredDomains;
					return {
						title: "",
						element: $compile("<arb-index featured-domains='::featuredDomains'></arb-index>")($rootScope),
					};
				}))
				.error(urlService.getErrorFunc("index"));
			}
		});
	});
}

app.service("urlService", function($http, $location, $ngSilentLocation, $rootScope, userService){
	var that = this;	
	
	var pageUpdater = null;
	this.setPageUpdater = function(value){
		pageUpdater = value;
	};
	var urlRules = [];
	this.when = function(urlPattern, rule) {
		var sections = urlPattern.split("/");
		var builder = [];
		var parameters = [];
		for (var n = 0; n < sections.length ; n++) {
			var section = sections[n];
			if (section == 0) {
				// ignore empty section
			}
			else if (section[0] == ":") {
				// Parameter capture
				parameters.push(section.substring(1));
				builder.push("\\/([^\\/]+)");
			}
			else if (section[0] == "?") {
				// Optional parameter capture
				parameters.push(section.substring(1));
				builder.push("(?:\\/([^\\/]+))?");
			}
			else {
				// match name
				builder.push("\\/"+section);
			}
		}
		builder.push("\\/");
		rule.urlPattern = new RegExp(builder.join(""));
		rule.parameters = parameters;
		urlRules.push(rule);
		return that;
	}
		
	this.resolveUrl = function() {
		// Get subdomain if any
		that.subdomain = undefined;
		var subdomainMatch = /^([A-Za-z0-9_]+)\.(localhost|arbital\.com)\/?$/.exec($location.host());
		if (subdomainMatch) {
			that.subdomain = subdomainMatch[1];
		}
		path = $location.path();
		var p = path.indexOf("?");
		if (p > -1) {
			if (!(p > 0 && path[p - 1] == "/")) {
				path = path.substring(0, p) + "/" + path.substring(p);
			}
		}
		else { 
			p = path.indexOf("#");
			if (p > -1) {
			if (!(p > 0 && path[p - 1] == "/")) {
				path = path.substring(0, p) + "/" + path.substring(p);
			}
			else if (!path.endsWith("/")) {
				path += "/";
			}
		}
		for (var ruleIndex = 0; ruleIndex < urlRules.length; ruleIndex++) {
			var rule = urlRules[ruleIndex];
			var m = rule.urlPattern.exec(url);
			if (m) {
				var args = {};
				var parameters = rule.parameters;
				for (var parameterIndex = 0; parameterIndex < parameters.length; parameterIndex++) {
					var parameter = parameters[parameterIndex];
					args[parameter] = m[parameterIndex + 1];
				}
				if (pageUpdater) {
					var name = rule.name;
					if (name) {
						if (pageUpdater(name, args)) {
							return; // The current page could handle the URL by modifying itself
						}
					}
				}
				pageUpdater = null;
				rule.handler(args);
				return;
			}
		}
	};
	this.loadingBarValue = 0;
	// Returns a function we can use as success handler for POST requests for dynamic data.
	// callback - returns {
	//   title: title to set for the window
	//   element: optional jQuery element to add dynamically to the body
	//   error: optional error message to print
	// }
	this.getSuccessFunc = function(callback) {
		return function(data) {
			// Sometimes we don't get data.
			if (data) {
				console.log("Dynamic request data:"); console.log(data);
				userService.processServerData(data);
				pageService.processServerData(data);
			}

			// Because the subdomain could have any case, we need to find the alias
			// in the loaded map so we can get the alias with correct case
			if (that.subdomain) {
				for (var pageAlias in pageService.pageMap) {
					if (that.subdomain.toUpperCase() === pageAlias.toUpperCase()) {
						that.subdomain = pageAlias;
						pageService.privateGroupId = pageService.pageMap[pageAlias].pageId;
						break;
					}
				}
			}

			// Get the results from page-specific callback
			$(".global-error").hide();
			var result = callback(data);
			if (result.error) {
				$(".global-error").text(result.error).show();
				document.title = "Error - Arbital";
			}
			if (result.element) {
				// Only show the element after it and all the children have been fully compiled and linked
				result.element.addClass("reveal-after-render-parent");
				var $loadingBar = $("#loading-bar");
				$loadingBar.show();
				that.loadingBarValue = 0;
				var startTime = (new Date()).getTime();
				var revealInterval = $interval(function() {
					var timePassed = ((new Date()).getTime() - startTime) / 1000;
					that.loadingBarValue = Math.min(100, timePassed * 30);
					var hiddenChildren = result.element.find(".reveal-after-render");
					if (hiddenChildren.length > 0) {
						hiddenChildren.each(function() {
							if ($(this).children().length > 0) {
								$(this).removeClass("reveal-after-render");
							}
						});
						return;
					}
					$interval.cancel(revealInterval);
					// Do short timeout to prevent some rendering bugs that occur on edit page
					$timeout(function() {
						result.element.removeClass("reveal-after-render-parent");
						$loadingBar.hide();
						$anchorScroll();
					}, 50);
				}, 50);

				$("[ng-view]").append(result.element);
			}

			$("body").toggleClass("body-fix", !result.removeBodyFix);

			if (result.title) {
				document.title = result.title + " - Arbital";
			}
		};
	};
	
	// Returns a function we can use as error handler for POST requests for dynamic data.
	this.getErrorFunc = function(urlPageType) {
		return function(data, status){
			console.error("Error /json/" + urlPageType + "/:"); console.log(data); console.log(status);
			$(".global-error").text(data).show();
			document.title = "Error - Arbital";
		};
	};
}
