'use strict';

import app from './angular.ts';
import {Editor} from './Markdown.Editor.ts';
import {getSanitizingConverter} from './Markdown.Sanitizer.ts';
import {InitMathjax} from './mathjax.ts';

// From demo-bundle.js
declare function loadAllDemos();

var notEscaped = '(^|\\\\`|\\\\\\[|(?:[^A-Za-z0-9_`[\\\\]|\\\\\\\\))';
var noParen = '(?=$|[^(])';
var nakedAliasMatch = '[\\-\\+]?[A-Za-z0-9_]+\\.?[A-Za-z0-9_]*';
export var aliasMatch = '(' + nakedAliasMatch + ')';
var anyUrlMatch = /\b((?:[a-z][\w-]+:(?:\/{1,3}|[a-z0-9%])|www\d{0,3}[.]|[a-z0-9.\-]+[.][a-z]{2,4}\/)(?:[^\s()<>]+|\(([^\s()<>]+|(\([^\s()<>]+\)))*\))+(?:\(([^\s()<>]+|(\([^\s()<>]+\)))*\)|[^\s`!()\[\]{};:'".,<>?«»“”‘’]))/i;

// [vote: alias]
export var voteEmbedRegexp = new RegExp(notEscaped +
		'\\[vote: ?' + aliasMatch + '\\]' + noParen, 'g');
// [alias/url text]
export var forwardLinkRegexp = new RegExp(notEscaped +
		'\\[([^ \\]]+?) (?![^\\]]*?\\\\\\])([^\\]]+?)\\]' + noParen, 'g');
// [alias]
export var simpleLinkRegexp = new RegExp(notEscaped +
		'\\[' + aliasMatch + '\\]' + noParen, 'g');
// [text](alias)
export var complexLinkRegexp = new RegExp(notEscaped +
		'\\[([^\\]]+?)\\]' + // match [Text]
		'\\(' + aliasMatch + '\\)', 'g'); // match (Alias)
// [@alias]
export var atAliasRegexp = new RegExp(notEscaped +
		'\\[@' + aliasMatch + '\\]' + noParen, 'g');

// markdownService provides a constructor you can use to create a markdown converter,
// either for converting markdown to text or editing.
app.service('markdownService', function($compile, $timeout, pageService, userService, urlService, stateService) {
	var that = this;

	// Store an array of page aliases that failed to load, so that we don't keep trying to reload them
	var failedPageAliases = {};

	// Trim + or - from beginning of the alias.
	var trimAlias = function(alias) {
		var firstAliasChar = alias.substring(0, 1);
		if (firstAliasChar == '-' || firstAliasChar == '+') {
			return alias.substring(1);
		}
		return alias;
	};

	// If prefix is '-', lowercase the first letter of text. Otherwise capitalize it.
	var getCasedText = function(text, prefix) {
		if (prefix == '-') {
			return text.substring(0, 1).toLowerCase() + text.substring(1);
		}
		return text.substring(0, 1).toUpperCase() + text.substring(1);
	};

	// Pass in a pageId to create an editor for that page
	var createConverterInternal = function(scope, pageId, isEditor = false) {
		// NOTE: not using $location, because we need port number
		var host = window.location.host;
		var converter = getSanitizingConverter();
		var editor = isEditor ? new Editor(converter, pageId) : undefined;
		var markdownPage = undefined;
		if (!isEditor && (pageId in stateService.pageMap)) {
			markdownPage = stateService.pageMap[pageId];
			markdownPage.todos = [];
			markdownPage.redAliases = {};
		}

		// Get the html for linking to the given alias. If alias is not found in page mape, the
		// link will be red.
		// text - optional text value for the url. If none is given, page's title will be used.
		//	If page is not found, page's alias will be used.
		var getLinkHtml = function(editor, alias, text = null) {
			var firstAliasChar = alias.substring(0, 1);
			var trimmedAlias = trimAlias(alias);
			var classText = 'intrasite-link';
			var page = stateService.pageMap[trimmedAlias];
			if (page) {
				var url = urlService.getPageUrl(page.pageId);
				if (!text) {
					text = getCasedText(page.title, firstAliasChar);
				}
				if (page.isDeleted) {
					classText += ' red-link';
				}
				return '<a href="' + url + '" class="' + classText + '" page-id="' + page.pageId + '">' +
					text + '</a>';
			}
			fetchLink(trimmedAlias, editor);
			classText += ' red-link';
			var url = urlService.getEditPageUrl(trimmedAlias);
			if (!text) {
				text = getCasedText(trimmedAlias, firstAliasChar).replace(/_/g, ' ');
			}
			if (!isEditor && markdownPage) {
				markdownPage.redAliases[alias] = text;
			}
			return '<a href="' + url + '" class="' + classText + '" page-id="">' + text + '</a>';
		};

		// Get info from BE to render the given page alias
		var fetchLink = function(pageAlias, editor) {
			if (!editor || pageAlias in failedPageAliases) return;
			// Try to load the page
			pageService.loadTitle(pageAlias, {
				silentFail: true,
				success: function() {
					if (pageAlias in stateService.pageMap) {
						editor.refreshPreview();
					} else {
						failedPageAliases[pageAlias] = true;
					}
				}
			});
		};

		// Process [summary(optional):markdown] blocks.
		var summaryBlockRegexp = new RegExp('^\\[summary(\\([^)\n\r]+\\))?: ?([\\s\\S]+?)\\] *(?=\Z|\n\Z|\n\n)', 'gm');
		converter.hooks.chain('preBlockGamut', function(text, runBlockGamut) {
			return text.replace(summaryBlockRegexp, function(whole, summaryName, summary) {
				if (isEditor) {
					return runBlockGamut('---\n\n**Summary' + (summaryName || '') + ':** ' + summary + '\n\n---');
				} else {
					return runBlockGamut('');
				}
			});
		});

		// Process %knows-requisite([alias]):markdown% blocks.
		var hasReqBlockRegexp = new RegExp('^(%+)(!?)knows-requisite\\(\\[' + aliasMatch + '\\]\\): ?([\\s\\S]+?)\\1 *(?=\Z|\n)', 'gm');
		converter.hooks.chain('preBlockGamut', function(text, runBlockGamut) {
			return text.replace(hasReqBlockRegexp, function(whole, bars, not, alias, markdown) {
				var pageId = (alias in stateService.pageMap) ? stateService.pageMap[alias].pageId : alias;
				var div = '<div ng-show=\'' + (not ? '!' : '') + 'arb.masteryService.hasMastery("' + pageId + '")\'>';
				if (isEditor) {
					div = '<div class=\'conditional-text editor-block\'>';
				}
				return div + runBlockGamut(markdown) + '\n\n</div>';
			});
		});

		// Process %wants-requisite([alias]):markdown% blocks.
		var wantsReqBlockRegexp = new RegExp('^(%+)(!?)wants-requisite\\(\\[' + aliasMatch + '\\]\\): ?([\\s\\S]+?)\\1 *(?=\Z|\n)', 'gm');
		converter.hooks.chain('preBlockGamut', function(text, runBlockGamut) {
			return text.replace(wantsReqBlockRegexp, function(whole, bars, not, alias, markdown) {
				var pageId = (alias in stateService.pageMap) ? stateService.pageMap[alias].pageId : alias;
				var div = '<div ng-show=\'' + (not ? '!' : '') + 'arb.masteryService.wantsMastery("' + pageId + '")\'>';
				if (isEditor) {
					div = '<div class=\'conditional-text editor-block\'>';
				}
				return div + runBlockGamut(markdown) + '\n\n</div>';
			});
		});

		// Process %if-before([alias]):markdown% blocks.
		var ifBeforeBlockRegexp = new RegExp('^(%+)(!?)if-(before|after)\\(\\[' + aliasMatch + '\\]\\): ?([\\s\\S]+?)\\1 *(?=\Z|\n)', 'gm');
		converter.hooks.chain('preBlockGamut', function(text, runBlockGamut) {
			return text.replace(ifBeforeBlockRegexp, function(whole, bars, not, beforeOrAfter, alias, markdown) {
				var pageId = (alias in stateService.pageMap) ? stateService.pageMap[alias].pageId : alias;
				var fnName = beforeOrAfter == 'before' ? 'isBefore' : 'isAfter';
				var div = '<div ng-show=\'' + (not ? '!' : '') + 'arb.pathService.' + fnName + '("' + pageId + '")\'>';
				if (isEditor) {
					div = '<div class=\'conditional-text editor-block\'>';
				}
				return div + runBlockGamut(markdown) + '\n\n</div>';
			});
		});

		// Process %todo:markdown% blocks.
		var todoBlockRegexp = new RegExp('^(%+)todo: ?([\\s\\S]+?)\\1 *(?=\Z|\n)', 'gm');
		converter.hooks.chain('preBlockGamut', function(text, runBlockGamut) {
			return text.replace(todoBlockRegexp, function(whole, bars, markdown) {
				if (isEditor) {
					return '<div class=\'todo-text editor-block\'>' + runBlockGamut(markdown) + '\n\n</div>';
				}
				markdownPage.todos.push(markdown);
				return '';
			});
		});

		// Process %fixme:markdown% blocks.
		var fixmeBlockRegexp = new RegExp('^(%+)fixme: ?([\\s\\S]+?)\\1 *(?=\Z|\n)', 'gm');
		converter.hooks.chain('preBlockGamut', function(text, runBlockGamut) {
			return text.replace(fixmeBlockRegexp, function(whole, bars, markdown) {
				if (isEditor) {
					return '<div class=\'fixme-text editor-block\'>' + runBlockGamut(markdown) + '\n\n</div>';
				}
				return '';
			});
		});

		// Process %comment:markdown% blocks.
		var commentBlockRegexp = new RegExp('^(%+)comment: ?([\\s\\S]+?)\\1 *(?=\Z|\n)', 'gm');
		converter.hooks.chain('preBlockGamut', function(text, runBlockGamut) {
			return text.replace(commentBlockRegexp, function(whole, bars, markdown) {
				if (isEditor) {
					return '<div class=\'info-text editor-block\'>' + runBlockGamut(markdown) + '\n\n</div>';
				}
				return '';
			});
		});

		// Process %hidden: text% blocks.
		var hiddenBlockRegexp = new RegExp('^(%+)hidden\\(([\\s\\S]+?)\\): ?([\\s\\S]+?)\\1 *(?=\Z|\n)', 'gm');
		converter.hooks.chain('preBlockGamut', function(text, runBlockGamut) {
			return text.replace(hiddenBlockRegexp, function(whole, bars, buttonText, text) {
				var blockText = runBlockGamut(text + '\n\n');
				var divClass = 'hidden-text';
				if (!isEditor) {
					divClass += ' display-none';
				} else {
					buttonText = '';
				}
				var html = '\n\n<div class=\'' + divClass + '\'>' + blockText + '\n\n</div>';
				return '<div arb-hidden-text button-text=\'' + buttonText + '\'>' + html + '</div>';
			});
		});

		// Process [multiple-choice(objectAlias): text
		// a: text
		// knows: [alias1],[alias2]...
		// wants: [alias1],[alias2]...
		// ] blocks.
		var mcBlockRegexp = new RegExp('^\\[multiple-choice\\(' + aliasMatch + '\\): ?([^\n]+?)\n' +
				'(a: ?[^\n]+?\n)' + // choice, e.g. "a: Carrots"
				'(knows: ?[^\n]+?\n)?' +
				'(wants: ?[^\n]+?\n)?' +
				'(-knows: ?[^\n]+?\n)?' +
				'(-wants: ?[^\n]+?\n)?' +
				'(path: ?[^\n]+?\n)?' +
				'(b: ?[^\n]+?\n)' + // choice, e.g. "b: Carrots"
				'(knows: ?[^\n]+?\n)?' +
				'(wants: ?[^\n]+?\n)?' +
				'(-knows: ?[^\n]+?\n)?' +
				'(-wants: ?[^\n]+?\n)?' +
				'(path: ?[^\n]+?\n)?' +
				'(c: ?[^\n]+?\n)?' + // choice, e.g. "c: Carrots"
				'(knows: ?[^\n]+?\n)?' +
				'(wants: ?[^\n]+?\n)?' +
				'(-knows: ?[^\n]+?\n)?' +
				'(-wants: ?[^\n]+?\n)?' +
				'(path: ?[^\n]+?\n)?' +
				'(d: ?[^\n]+?\n)?' + // choice, e.g. "d: Carrots"
				'(knows: ?[^\n]+?\n)?' +
				'(wants: ?[^\n]+?\n)?' +
				'(-knows: ?[^\n]+?\n)?' +
				'(-wants: ?[^\n]+?\n)?' +
				'(path: ?[^\n]+?\n)?' +
				'(e: ?[^\n]+?\n)?' + // choice, e.g. "e: Carrots"
				'(knows: ?[^\n]+?\n)?' +
				'(wants: ?[^\n]+?\n)?' +
				'(-knows: ?[^\n]+?\n)?' +
				'(-wants: ?[^\n]+?\n)?' +
				'(path: ?[^\n]+?\n)?' +
				'\\] *(?=\Z|\n)', 'gm');
		converter.hooks.chain('preBlockGamut', function(text, runBlockGamut) {
			return text.replace(mcBlockRegexp, function() {
				var result = [];
				// Process captured groups
				for (var n = 2; n < arguments.length; n++) {
					var arg = arguments[n];
					if (+arg) break; // there are extra arguments that we don't need, starting with some number
					if (!arg) continue;
					if (n == 2) { // question text
						result.push(arg + '\n\n');
					} else {
						// Match answer line
						var match = arg.match(/^([a-e]): ?([\s\S]+?)\n$/);
						if (match) {
							result.push('- ' + match[2] + '\n');
							continue;
						}
						result.push(' - ' + arg);
					}
				}
				return '<arb-multiple-choice page-id=\'' + pageId + '\' object-alias=\'' + arguments[1] + '\'>' +
					runBlockGamut(result.join('')) + '\n\n</arb-multiple-choice>';
			});
		});

		// Process [checkbox(alias): text
		// y:
		// knows: [alias1],[alias2]...
		// wants: [alias1],[alias2]...
		// ] blocks.
		var checkboxBlockRegexp = new RegExp('^\\[checkbox\\(' + aliasMatch + '\\): ?([^\n]+?)\n' +
				'(y!?:\n)' +
				'(knows: ?[^\n]+?\n)?' +
				'(wants: ?[^\n]+?\n)?' +
				'(-knows: ?[^\n]+?\n)?' +
				'(-wants: ?[^\n]+?\n)?' +
				'(path: ?[^\n]+?\n)?' +
				'(n!?:\n)' +
				'(knows: ?[^\n]+?\n)?' +
				'(wants: ?[^\n]+?\n)?' +
				'(-knows: ?[^\n]+?\n)?' +
				'(-wants: ?[^\n]+?\n)?' +
				'(path: ?[^\n]+?\n)?' +
				'\\] *(?=\Z|\n)', 'gm');
		converter.hooks.chain('preBlockGamut', function(text, runBlockGamut) {
			return text.replace(checkboxBlockRegexp, function() {
				var defaultValue = 'n';
				var result = [];
				// Process captured groups
				for (var n = 2; n < arguments.length; n++) {
					var arg = arguments[n];
					if (+arg) break; // there are extra arguments that we don't need, starting with some number
					if (!arg) continue;
					if (n == 2) { // checkbox text
						result.push(arg + '\n\n');
					} else {
						// Match answer line
						var match = arg.match(/^([yn]!?):\n$/);
						if (match) {
							var answer = match[1];
							if (answer.length >= 1 && answer[1] == '!') {
								defaultValue = answer[0];
							}
							result.push('- ' + answer + '\n');
							continue;
						}
						result.push(' - ' + arg);
					}
				}
				return '<arb-checkbox page-id=\'' + pageId +
					'\' object-alias=\'' + arguments[1] +
					'\' default=\'' + defaultValue + '\'>' +
					runBlockGamut(result.join('')) + '\n\n</arb-checkbox>';
			});
		});

		// Process [toc:] block.
		var tocBlockRegexp = new RegExp('^\\[toc:\\] *(?=\Z|\n\Z|\n\n)', 'gm');
		converter.hooks.chain('preBlockGamut', function(text, runBlockGamut) {
			return text.replace(tocBlockRegexp, function(whole) {
				return '<arb-table-of-contents page-id=\'' + pageId + '\'></arb-table-of-contents>';
			});
		});

		// Process [visualization(log-graph-demo):] block.
		var vizBlockRegexp = new RegExp('^\\[visualization\\(([^)]+)\\):\\] *(?=\Z|\n\Z|\n\n)', 'gm');
		converter.hooks.chain('preBlockGamut', function(text, runBlockGamut) {
			return text.replace(vizBlockRegexp, function(whole, name) {
				return '<div class=\'react-demo\' data-demo-name="' + name + '">\n\n</div>';
			});
		});

		// Mathjax
		var getMathjaxRegexp = function(isBlock) {
			var regexp = '';
			if (isBlock) {
				regexp += '[\\s\\S]*?';
			} else {
				regexp += '[^\n]*?'; // spans don't include new line
			}
			regexp += '[^\\\\]';
			return regexp;
		};
		// Process $$$mathjax$$$ spans (when you need to embed $...$ within)
		var mathjaxSpan3Regexp = new RegExp(notEscaped + '(~D~D~D' + getMathjaxRegexp(false) + '~D~D~D)', 'g');
		converter.hooks.chain('preSpanGamut', function(text) {
			return text.replace(mathjaxSpan3Regexp, function(whole, prefix, mathjaxText) {
				var encodedText = encodeURIComponent(mathjaxText.substring(6, mathjaxText.length - 6));
				var key = '$' + encodedText + '$';
				var cachedValue = stateService.getMathjaxCacheValue(key);
				var style = cachedValue ? ('style=\'' + cachedValue.style + ';display:inline-block;\' ') : '';
				return prefix + '<span ' + style + 'arb-math-compiler="' + key + '">&nbsp;</span>';
			});
		});
		// Process $$mathjax$$ spans.
		var mathjaxSpan2Regexp = new RegExp(notEscaped + '(~D~D' + getMathjaxRegexp(true) + '~D~D)', 'g');
		converter.hooks.chain('preSpanGamut', function(text) {
			return text.replace(mathjaxSpan2Regexp, function(whole, prefix, mathjaxText) {
				if (mathjaxText.substring(0, 6) == '~D~D~D') return whole;
				var encodedText = encodeURIComponent(mathjaxText.substring(4, mathjaxText.length - 4));
				var key = '$$' + encodedText + '$$';
				var cachedValue = stateService.getMathjaxCacheValue(key);
				var style = cachedValue ? ('style=\'' + cachedValue.style + '\' ') : '';
				return prefix + '<span ' + style + 'class=\'mathjax-div\' arb-math-compiler="' + key + '">&nbsp;</span>';
			});
		});
		// Process $mathjax$ spans.
		var mathjaxSpanRegexp = new RegExp('(^|\\s|"|\'|\\(|\\[|-|_|>)(~D' + getMathjaxRegexp(false) + '~D)', 'g');
		converter.hooks.chain('preSpanGamut', function(text) {
			return text.replace(mathjaxSpanRegexp, function(whole, prefix, mathjaxText) {
				if (mathjaxText.substring(0, 4) == '~D~D') return whole;
				var encodedText = encodeURIComponent(mathjaxText.substring(2, mathjaxText.length - 2));
				var key = '$' + encodedText + '$';
				var cachedValue = stateService.getMathjaxCacheValue(key);
				var style = cachedValue ? ('style=\'' + cachedValue.style + ';display:inline-block;\' ') : '';
				return prefix + '<span ' + style + 'arb-math-compiler="' + key + '">&nbsp;</span>';
			});
		});

		// Process %note: markdown% spans.
		var noteSpanRegexp = new RegExp(notEscaped + '(%+)note: ?([\\s\\S]+?)\\2', 'g');
		converter.hooks.chain('preSpanGamut', function(text) {
			return text.replace(noteSpanRegexp, function(whole, prefix, bars, markdown) {
				if (isEditor) {
					return prefix + '<span class=\'conditional-text\'>' + markdown + '</span>';
				}
				return prefix + '<span class=\'markdown-note\' arb-text-popover-anchor>' + markdown + '</span>';
			});
		});

		// Process %knows-requisite([alias]): markdown% spans.
		var hasReqSpanRegexp = new RegExp(notEscaped + '(%+)(!?)knows-requisite\\(\\[' + aliasMatch + '\\]\\): ?([\\s\\S]+?)\\2', 'g');
		converter.hooks.chain('preSpanGamut', function(text) {
			return text.replace(hasReqSpanRegexp, function(whole, prefix, bars, not, alias, markdown) {
				var pageId = (alias in stateService.pageMap) ? stateService.pageMap[alias].pageId : alias;
				var span = '<span ng-show=\'' + (not ? '!' : '') + 'arb.masteryService.hasMastery("' + pageId + '")\'>';
				if (isEditor) {
					span = '<span class=\'conditional-text\'>';
				}
				return prefix + span + markdown + '</span>';
			});
		});

		// Process %wants-requisite([alias]): markdown% spans.
		var wantsReqSpanRegexp = new RegExp(notEscaped + '(%+)(!?)wants-requisite\\(\\[' + aliasMatch + '\\]\\): ?([\\s\\S]+?)\\2', 'g');
		converter.hooks.chain('preSpanGamut', function(text, run) {
			return text.replace(wantsReqSpanRegexp, function(whole, prefix, bars, not, alias, markdown) {
				var pageId = (alias in stateService.pageMap) ? stateService.pageMap[alias].pageId : alias;
				var span = '<span ng-show=\'' + (not ? '!' : '') + 'arb.masteryService.wantsMastery("' + pageId + '")\'>';
				if (isEditor) {
					span = '<span class=\'conditional-text\'>';
				}
				return prefix + span + markdown + '</span>';
			});
		});

		// Process %if-before([alias]): markdown% spans.
		var ifBeforeSpanRegexp = new RegExp(notEscaped + '(%+)(!?)if-(before|after)\\(\\[' + aliasMatch + '\\]\\): ?([\\s\\S]+?)\\2', 'g');
		converter.hooks.chain('preSpanGamut', function(text) {
			return text.replace(ifBeforeSpanRegexp, function(whole, prefix, bars, not, beforeOrAfter, alias, markdown) {
				var pageId = (alias in stateService.pageMap) ? stateService.pageMap[alias].pageId : alias;
				var fnName = beforeOrAfter == 'before' ? 'isBefore' : 'isAfter';
				var span = '<span ng-show=\'' + (not ? '!' : '') + 'arb.pathService.' + fnName + '("' + pageId + '")\'>';
				if (isEditor) {
					span = '<span class=\'conditional-text\'>';
				}
				return prefix + span + markdown + '</span>';
			});
		});

		// Process [todo:text] spans.
		var todoSpanRegexp = new RegExp(notEscaped + '\\[todo: ?([^\\]]+?)\\]' + noParen, 'g');
		converter.hooks.chain('preSpanGamut', function(text) {
			return text.replace(todoSpanRegexp, function(whole, prefix, text) {
				if (isEditor) {
					return prefix + '<span class=\'todo-text\'>' + text + '</span>';
				}
				markdownPage.todos.push(text);
				return prefix;
			});
		});

		// Process [fixme:text] spans.
		var fixmeSpanRegexp = new RegExp(notEscaped + '\\[fixme: ?([^\\]]+?)\\]' + noParen, 'g');
		converter.hooks.chain('preSpanGamut', function(text) {
			return text.replace(fixmeSpanRegexp, function(whole, prefix, text) {
				if (isEditor) {
					return prefix + '<span class=\'fixme-text\'>' + text + '</span>';
				}
				return prefix;
			});
		});

		// Process [comment:text] spans.
		var commentSpanRegexp = new RegExp(notEscaped +
				'\\[comment: ?([^\\]]+?)\\]' + noParen, 'g');
		converter.hooks.chain('preSpanGamut', function(text) {
			return text.replace(commentSpanRegexp, function(whole, prefix, text) {
				if (isEditor) {
					return prefix + '<span class=\'info-text\'>';
				}
				return prefix;
			});
		});

		// Process [vote:alias] spans.
		converter.hooks.chain('preSpanGamut', function(text) {
			return text.replace(voteEmbedRegexp, function(whole, prefix, alias) {
				return prefix + '[Embedded ' + alias + ' vote. ](' + urlService.getPageUrl(alias) + '/?embedVote=1)';
			});
		});

		// Convert [ text] spans into links.
		var spaceTextRegexp = new RegExp(notEscaped +
				'\\[ ([^\\]]+?)\\]' + noParen, 'g');
		converter.hooks.chain('preSpanGamut', function(text) {
			return text.replace(spaceTextRegexp, function(whole, prefix, text) {
				var url = urlService.getNewPageUrl();
				return prefix + '<a href="' + url + '" class="intrasite-link red-link" page-id="">' +
					text + '</a>';
			});
		});

		// Convert [alias/url text] spans into links.
		converter.hooks.chain('preSpanGamut', function(text) {
			return text.replace(forwardLinkRegexp, function(whole, prefix, alias, text) {
				var matches = alias.match(anyUrlMatch);
				if (matches) {
					// This is just a normal url.
					return prefix + '[' + text + '](' + matches[0] + ')';
				}
				matches = alias.match(aliasMatch);
				if (!matches || matches[0] != alias) {
					// No alias match
					return whole;
				}
				return prefix + getLinkHtml(editor, alias, text);
			});
		});

		// Convert [alias] spans into links.
		converter.hooks.chain('preSpanGamut', function(text) {
			return text.replace(simpleLinkRegexp, function(whole, prefix, alias) {
				return prefix + getLinkHtml(editor, alias);
			});
		});

		// Convert [@alias] spans into links.
		converter.hooks.chain('preSpanGamut', function(text) {
			return text.replace(atAliasRegexp, function(whole, prefix, alias) {
				var html = getLinkHtml(editor, alias);
				return prefix + html.replace('page-id', 'user-id').replace('intrasite-link', 'user-link');
			});
		});

		if (isEditor) {
			// Setup the editor stuff.
			if (!userService.user.ignoreMathjax) {
				InitMathjax(converter, editor, pageId);
			}
			var $wmdPreview = $('#wmd-preview' + pageId);
			converter.hooks.chain('postConversion', function(text) {
				$timeout(function() {
					that.processLinks(scope, $wmdPreview);
					that.compileChildren(scope, $wmdPreview, {isEditor: true});
				});
				return text;
			});

			editor.run();
		} else {
			InitMathjax(converter);
		}

		return converter;
	};

	// Process all the links in the give element.
	this.processLinks = function(scope, $pageText, isEditor) {
		var index = 1; // start with 1, because masteryMap is at 0 (see pageService.js)
		$pageText.find('arb-multiple-choice,arb-checkbox').each(function() {
			$(this).attr('index', index++);
		});
	};

	// Compile all the children of arb-markdown
	// options = {
	//	isEditor: this compilation is done for a page preview
	//	skipCompile: don't compile the children
	// }
	this.compileChildren = function(scope, $pageText, options) {
		options = options || {};

		// NOTE: have to compile children individually because otherwise there is a bug
		// with intrasite popovers in preview.
		if (!options.skipCompile) {
			$pageText.children().each(function(index) {
				$compile($(this))(scope);
			});
			loadAllDemos();
		}

		// If first time around, set up the functions
		if (scope._cacheMath === undefined) {
			// Go through all rendered mathjax elements and cache the HTML
			scope._cacheMath = function(elements) {
				for (var n = 0; n < elements.length; n++) {
					var $element = elements[n].$element;

					// Check that the element is still in DOM and the contents were rendered
					var $contentElement = $element.find('.MathJax, .MathJax_Display, .MathJax_SVG, .MathJax_SVG_Display');
					if ($element.closest('body').length <= 0 || $contentElement.length <= 0) {
						continue;
					}

					stateService.cacheMathjax(elements[n].encodedMathjaxText, {
						html: $element.html(),
						style: 'width:' + $contentElement.width() + 'px;' +
							'height:' + $contentElement.height() + 'px',
					});
				}
			};

			scope._restartMathJax = function() {
				MathJax.Hub.cancelTypeset = false;
			};
		}

		if (options.isEditor) {
			// Delay all math rendering to prevent constant flickering when typing
			$timeout.cancel(scope._mathRenderPromise);
			scope._mathRenderPromise = $timeout(function() {
				// Track all elements that should be cached after they are rendered
				var elements = [];
				// Go through all mathjax elements
				$pageText.find('[arb-math-compiler]').each(function() {
					var $element = $(this);
					var encodedMathjaxText = $element.attr('arb-math-compiler');
					// Try to read from cache
					var cachedValue = stateService.getMathjaxCacheValue(encodedMathjaxText);
					if (cachedValue) {
						$timeout(function() {
							$element.html(cachedValue.html);
						});
					} else if ($element.text().length <= 1) {
						$element.text('$~' + decodeURIComponent(encodedMathjaxText) + '~$');
						elements.push({'$element': $element, encodedMathjaxText: encodedMathjaxText});
					}
				});
				MathJax.Hub.Cancel();
				MathJax.Hub.Queue(['_restartMathJax', scope]);
				MathJax.Hub.Queue(['Typeset', MathJax.Hub, $pageText.get(0)]);
				MathJax.Hub.Queue(['_cacheMath', scope, elements]);
			}, scope._mathRenderPromise ? 500 : 0);
		} else {
			// Compile all mathjax at once
			$pageText.find('[arb-math-compiler]').each(function() {
				var $element = $(this);
				var encodedMathjaxText = '$~' + $element.attr('arb-math-compiler') + '~$';
				$element.text(decodeURIComponent(encodedMathjaxText));
			});
			MathJax.Hub.Queue(['Typeset', MathJax.Hub, $pageText.get(0)]);
		}
	};

	this.createConverter = function(scope, pageId) {
		return createConverterInternal(scope, pageId);
	};

	this.createEditConverter = function(scope, pageId) {
		failedPageAliases = {};
		return createConverterInternal(scope, pageId, true);
	};
});

