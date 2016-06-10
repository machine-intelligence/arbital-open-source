'use strict';

// Directive for rendering markdown text.
app.directive('arbMarkdown', function($compile, $timeout, arb) {
	return {
		scope: {
			// One of these ids has to be set.
			pageId: '@',
			markId: '@',

			// If summary name is set, we'll display the page's corresponding summary
			summaryName: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = !!$scope.pageId ? arb.stateService.pageMap[$scope.pageId] : undefined;
			$scope.mark = !!$scope.markId ? arb.markService.markMap[$scope.markId] : undefined;
		},
		link: function(scope, element, attrs) {
			element.addClass('markdown-text reveal-after-render');

			// Convert page text to html.
			// Note: converter takes pageId, which might not be set if we are displaying
			// a mark, but it should be ok, since the mark doesn't have most markdown features.
			var converter = arb.markdownService.createConverter(scope, scope.pageId);
			var text = '';
			if (scope.page) {
				if (scope.summaryName) {
					text = scope.page.summaries[scope.summaryName];
					text = text || scope.page.summaries['Summary']; // jscs:ignore requireDotNotation
					if (!text) {
						// Take the first one.
						for (var key in scope.page.summaries) {
							text = scope.page.summaries[key];
							break;
						}
					}
				} else {
					text = scope.page.text;
					if (scope.page.anchorText) {
						text = '>' + scope.page.anchorText + '\n\n' + text;
					}
				}
			} else if (scope.mark) {
				text = scope.mark.anchorContext;
			}

			var html = converter.makeHtml(text);
			var $pageText = element;
			$pageText.html(html);
			$timeout(function() {
				arb.markdownService.processLinks(scope, $pageText);
				element.removeClass('reveal-after-render');
				$timeout(function() {
					//MathJax.Hub.Queue(['Typeset', MathJax.Hub, $pageText.get(0)]);
					//MathJax.Hub.Queue(['compileChildren', arb.markdownService, scope, $pageText]);
					arb.markdownService.compileChildren(scope, $pageText);
					// Highlight the anchorText for marks.
					MathJax.Hub.Queue(function() {
						if (scope.mark) {
							var highlightClass = 'inline-comment-highlight-hover';
							createInlineCommentHighlight(element.children().get(0), scope.mark.anchorOffset,
								scope.mark.anchorOffset + scope.mark.anchorText.length, highlightClass);
						}
					});
				});
			});
		},
	};
});
