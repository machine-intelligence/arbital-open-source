'use strict';

// Directive for table of contents
app.directive('arbTableOfContents', function($timeout, $http, $compile, pageService, userService) {
	return {
		templateUrl: 'static/html/tableOfContents.html',
		transclude: true,
		scope: {
			pageId: '@',
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.showToc = true;
			$scope.toc = [];
		},
		link: function(scope, element, attrs) {
			var $parent = element.closest('arb-markdown');

			// We compute sections in a bit of a weird way because we can have H3 or H2 come
			// before we have H1, so we need to count them as whole sections until we find H1,
			// after which they become subsections.
			var counts = [{count: 0, minHeader: 4}];

			// Add a row to TOC given current state.
			var addContentRow = function(header) {
				var section = '' + counts[0].count;
				if (counts.length > 1) {
					section += '.' + counts[1].count;
				}
				if (counts.length > 2) {
					section += '.' + counts[2].count;
				}
				var id = 'h-' + scope.pageId + '-' + section;
				var url = '#' + id;
				var row = {section: section, header: header, tabSize: counts.length - 1, url: url};
				scope.toc.push(row);
				return id;
			};

			// Go through all the headers and create TOC entries
			$parent.find('h1,h2,h3').each(function() {
				var $this = $(this);
				var headerType = $this.prop('nodeName');
				var currentHeader;
				if (headerType === 'H1') {
					currentHeader = 1;
				} else if (headerType === 'H2') {
					currentHeader = 2;
				} else if (headerType === 'H3') {
					currentHeader = 3;
				} else {
					return;
				}
				var smallestIndex = -1;
				for (var n = 0; n < counts.length; n++) {
					if (currentHeader <= counts[n].minHeader) {
						counts[n].count++;
						counts[n].minHeader = currentHeader;
						break;
					}
				}
				if (n >= counts.length) {
					// Didn't find a count we could replace, so append one
					counts.push({count: 1, minHeader: currentHeader});
				} else {
					counts = counts.slice(0, n + 1);
				}
				var id = addContentRow($this.text());
				$this.attr('id', id);
			});
		},
	};
});

