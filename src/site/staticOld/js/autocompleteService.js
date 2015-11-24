"use strict";

// Autocomplete service provides data for autocompletion.
app.service("autocompleteService", function($http, $compile, pageService){
	var that = this;

	// Set how to render search results for the given autocomplete input.
	this.setAutocompleteRendering = function($input, scope, resultsAreLinks) {
		$input.data("ui-autocomplete")._renderItem = function(ul, item) {
			var elementType = "span";
			var elementTypeEnd = "span";
			if (resultsAreLinks) {
				elementType = "a href='" + pageService.getPageUrl(item.label) + "'";
				elementTypeEnd = "a";
			}
			var $el = $compile("<li class='ui-menu-item'><" + elementType +
				" arb-likes-page-title class='search-result' page-id='" + item.value +
				"' show-clickbait='true' is-search-result='true'></" + elementTypeEnd + "></li>")(scope);
			$el.attr("data-value", item.value);
			return $el.appendTo(ul);
		};
	};

	// Take data we get from BE search, and extract the data to forward it to
	// an autocompelete input. Also update the pageMap.
	this.processAutocompleteResults = function(data) {
		if (!data) return [];
		// Add new pages to the pageMap.
		for (var pageId in data.pages) {
			pageService.addPageToMap(data.pages[pageId]);
		}
		// Create list of results we can give to autocomplete.
		var resultList = [];
		var hits = data.result.search.hits;
		for (var n = 0; n < hits.length; n++) {
			var source = hits[n]._source;
			resultList.push({
				value: source.alias,
				label: source.pageId,
				alias: source.alias,
				title: source.title,
				clickbait: source.clickbait,
				seeGroupId: source.seeGroupId,
				score: hits[n]._score,
			});
		}
		return resultList;
	};


	// Do a normal search with the given options.
	// options = {
	//	term: string to search for
	//	pageType: contraint for what type of pages we are looking for
	// }
	// Returns: list of results
	this.performSearch = function(options, callback) {
		$http({method: "POST", url: "/json/search/", data: JSON.stringify(options)})
			.success(function(data, status){
				var results = that.processAutocompleteResults(data);
				if (callback) callback(results);
			})
			.error(function(data, status){
				console.log("Error loading /search/ autocomplete data:"); console.log(data); console.log(status);
				if (callback) callback({});
			});
	}

	// Load data for autocompleting parents search.
	this.parentsSource = function(request, callback) {
		$http({method: "POST", url: "/json/parentsSearch/", data: JSON.stringify(request)})
			.success(function(data, status){
				var results = that.processAutocompleteResults(data);
				if (callback) callback(results);
			})
			.error(function(data, status){
				console.log("Error loading /parentsSearch/ autocomplete data:"); console.log(data); console.log(status);
				callback([]);
			});
	};

	// Load data for autocompleting user search.
	this.userSource = function(request, callback) {
		$http({method: "POST", url: "/json/userSearch/", data: JSON.stringify(request)})
			.success(function(data, status){
				var results = that.processAutocompleteResults(data);
				if (callback) callback(results);
			})
			.error(function(data, status){
				console.log("Error loading /userSearch/ autocomplete data:"); console.log(data); console.log(status);
				callback([]);
			});
	};

	// Set up autocompletion based on parents search for the given input field.
	this.setupParentsAutocomplete = function($input, selectCallback) {
	  $input.autocomplete({
			source: that.parentsSource,
			minLength: 3,
			delay: 300,
			focus: function (event, ui) {
				return false;
			},
			select: function (event, ui) {
				return selectCallback(event, ui);
			}
	  });
	}

	// Find other pages similar to the page with the given data.
	this.findSimilarPages = function(pageData, callback) {
		$http({method: "POST", url: "/json/similarPageSearch/", data: JSON.stringify(pageData)})
		.success(function(data, status){
			var results = that.processAutocompleteResults(data);
			if (callback) callback(results);
		})
		.error(function(data, status){
			console.log("Error doing similar page search:"); console.log(data); console.log(status);
		});
	};
});
