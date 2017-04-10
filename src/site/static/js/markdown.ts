'use strict';

import app from './angular.ts';
import {createInlineCommentHighlight} from './inlineCommentUtil.ts';

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
		link: function(scope: any, element, attrs) {
			var doEverything = function() {
				element.addClass('markdown-text reveal-after-render invisible');
				var $progressBar = $compile('<md-progress-linear md-mode="query"></md-progress-linear>')(scope);
				$progressBar.insertBefore(element);

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

				// Called when the element is ready to be shown.
				var showElement = function() {
					element.removeClass('invisible');
					$progressBar.remove();
				};

				var html = converter.makeHtml(text);
				var $pageText = element;
				$pageText.html(html);
				$timeout(function() {
					arb.markdownService.processLinks(scope, $pageText);
					element.removeClass('reveal-after-render');
					$timeout(function() {
						arb.markdownService.compileChildren(scope, $pageText);
						// Highlight the anchorText for marks.
						MathJax.Hub.Queue(function() {
							showElement();
							if (scope.mark) {
								var highlightClass = 'inline-comment-highlight-hover';
								createInlineCommentHighlight(element.children().get(0), scope.mark.anchorOffset,
									scope.mark.anchorOffset + scope.mark.anchorText.length, highlightClass);
							}
						});
						// In case Mathjax fails or something, remove the invisible class after a delay
						$timeout(function() {
							showElement();
						}, 1500);
					});
				});
			};
			doEverything();

			scope.$watch(function() {return scope.page.text;}, function() {
				doEverything();
			});
		},
	};
});
