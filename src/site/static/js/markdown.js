'use strict';

var notEscaped = '(^|\\\\`|\\\\\\[|(?:[^A-Za-z0-9_`[\\\\]|\\\\\\\\))';
var noParen = '(?=$|[^(])';
var nakedAliasMatch = '[\\-\\+]?[A-Za-z0-9_]+\\.?[A-Za-z0-9_]*';
var aliasMatch = '(' + nakedAliasMatch + ')';
var anyUrlMatch = /\b((?:[a-z][\w-]+:(?:\/{1,3}|[a-z0-9%])|www\d{0,3}[.]|[a-z0-9.\-]+[.][a-z]{2,4}\/)(?:[^\s()<>]+|\(([^\s()<>]+|(\([^\s()<>]+\)))*\))+(?:\(([^\s()<>]+|(\([^\s()<>]+\)))*\)|[^\s`!()\[\]{};:'".,<>?«»“”‘’]))/i;

// [vote: alias]
var voteEmbedRegexp = new RegExp(notEscaped +
		'\\[vote: ?' + aliasMatch + '\\]' + noParen, 'g');
// [alias/url text]
var forwardLinkRegexp = new RegExp(notEscaped +
		'\\[([^ \\]]+?) (?![^\\]]*?\\\\\\])([^\\]]+?)\\]' + noParen, 'g');
// [alias]
var simpleLinkRegexp = new RegExp(notEscaped +
		'\\[' + aliasMatch + '\\]' + noParen, 'g');
// [text](alias)
var complexLinkRegexp = new RegExp(notEscaped +
		'\\[([^\\]]+?)\\]' + // match [Text]
		'\\(' + aliasMatch + '\\)', 'g'); // match (Alias)
// [@alias]
var atAliasRegexp = new RegExp(notEscaped +
		'\\[@' + aliasMatch + '\\]' + noParen, 'g');

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
			$scope.page = !!$scope.pageId ? arb.pageService.pageMap[$scope.pageId] : undefined;
			$scope.mark = !!$scope.markId ? arb.pageService.markMap[$scope.markId] : undefined;
		},
		link: function(scope, element, attrs) {
			element.addClass('markdown-text reveal-after-render');

			// Remove the class manually in case the text is empty
			$timeout(function() {
				element.removeClass('reveal-after-render');
			});

			// Convert page text to html.
			// Note: converter takes pageId, which might not be set if we are displaying
			// a mark, but it should be ok, since the mark doesn't have most markdown features.
			var converter = arb.markdownService.createConverter(scope.pageId);
			if (scope.page) {
				var html =  scope.page.text;
				if (scope.page.anchorText) {
					html = '>' + scope.page.anchorText + '\n\n' + html;
				}
			} else if (scope.mark) {
				var html = scope.mark.anchorContext;
			}
			if (scope.summaryName) {
				html = scope.page.summaries[scope.summaryName];
				html = html || scope.page.summaries['Summary']; // jscs:ignore requireDotNotation
				if (!html) {
					// Take the first one.
					for (var key in scope.page.summaries) {
						html = scope.page.summaries[key];
						break;
					}
				}
			}
			var html = converter.makeHtml(html);
			var $pageText = element;
			$pageText.html(html);
			window.setTimeout(function() {
				MathJax.Hub.Queue(['Typeset', MathJax.Hub, $pageText.get(0)]);
				MathJax.Hub.Queue(['processLinks', arb.markdownService, scope, $pageText]);
				// Highlight the anchorText for marks.
				MathJax.Hub.Queue(function() {
					if (scope.mark) {
						var highlightClass = 'inline-comment-highlight-hover';
						createInlineCommentHighlight(element.children().get(0), scope.mark.anchorOffset,
							scope.mark.anchorOffset + scope.mark.anchorText.length, highlightClass);
					}
				});
			});
		},
	};
});
