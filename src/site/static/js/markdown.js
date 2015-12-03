"use strict";

var notEscaped = "(^|\\\\`|\\\\\\[| |>|\n)";
var noParen = "(?=$|[^(])";
var aliasMatch = "([A-Za-z0-9]+\\.?[A-Za-z0-9]*)";
// [vote: alias]
var voteEmbedRegexp = new RegExp(notEscaped + 
		"\\[vote: ?" + aliasMatch + "\\]" + noParen, "g");
// [alias/url text] 
var forwardLinkRegexp = new RegExp(notEscaped + 
		"\\[([^ \\]]+?) ([^\\]]+?)\\]" + noParen, "g");
// [alias]
var simpleLinkRegexp = new RegExp(notEscaped + 
		"\\[" + aliasMatch + "\\]" + noParen, "g");
// [text](alias)
var complexLinkRegexp = new RegExp(notEscaped + 
		"\\[([^\\]]+?)\\]" + // match [Text]
		"\\(" + aliasMatch + "\\)", "g"); // match (Alias)
// [text](url)
var urlLinkRegexp = new RegExp(notEscaped + 
		"\\[([^\\]]+?)\\]" + // match [Text]
		"\\((http://" + RegExp.escape(window.location.host) + "/pages/)" + aliasMatch + "\\)", "g"); // match (Url)
// [@alias]
var atAliasRegexp = new RegExp(notEscaped + 
		"\\[@" + aliasMatch + "\\]" + noParen, "g");

// markdownFactory provides a constructor you can use to create a markdown converter,
// either for converting markdown to text or editing.
app.factory("markdownFactory", function(pageService){
	// Pass in a pageId to create an editor for that page
	return function(pageId) {
		var host = window.location.host;
		this.converter = Markdown.getSanitizingConverter();

		// Process [todo:text] spans.
		var todoSpanRegexp = new RegExp(notEscaped + 
				"\\[todo: ?([^\\]]+?)\\]" + noParen, "g");
		this.converter.hooks.chain("preSpanGamut", function (text) {
			return text.replace(todoSpanRegexp, function (whole, prefix, alias) {
				return prefix;
			});
		});

		// Process [comment:text] spans.
		var commentSpanRegexp = new RegExp(notEscaped + 
				"\\[comment: ?([^\\]]+?)\\]" + noParen, "g");
		this.converter.hooks.chain("preSpanGamut", function (text) {
			return text.replace(commentSpanRegexp, function (whole, prefix, alias) {
				return prefix;
			});
		});

		// Process [vote:alias] spans.
		this.converter.hooks.chain("preSpanGamut", function (text) {
			return text.replace(voteEmbedRegexp, function (whole, prefix, alias) {
				return prefix + "[Embedded " + alias + " vote. ](http://" + host + "/pages/" + alias + "/?embedVote=1)";
			});
		});

		// Convert [ text] spans into links.
		var spaceTextRegexp = new RegExp(notEscaped + 
				"\\[ ([^\\]]+?)\\]" + noParen, "g");
		this.converter.hooks.chain("preSpanGamut", function (text) {
			return text.replace(spaceTextRegexp, function (whole, prefix, text) {
				return prefix + "[" + text + "](" + "0" + ")";
			});
		});

		// Convert [alias/url text] spans into links.
		this.converter.hooks.chain("preSpanGamut", function (text) {
			return text.replace(forwardLinkRegexp, function (whole, prefix, alias, text) {
				return prefix + "[" + text + "](" + alias + ")";
			});
		});

		// Convert [alias] spans into links.
		this.converter.hooks.chain("preSpanGamut", function (text) {
			return text.replace(simpleLinkRegexp, function (whole, prefix, alias) {
				var page = pageService.pageMap[alias];
				if (page) {
					var url = "http://" + host + "/pages/" + page.pageId;
					return prefix + "[" + page.title + "](" + url + ")";
				} else {
					return prefix + "[" + alias + "](" + alias + ")";
				}
			});
		});

		// Convert [@alias] spans into links.
		this.converter.hooks.chain("preSpanGamut", function (text) {
			return text.replace(atAliasRegexp, function (whole, prefix, alias) {
				var page = pageService.pageMap[alias];
				if (page) {
					var url = "http://" + host + "/user/" + page.pageId + "/";
					return prefix + "[" + page.title + "](" + url + ")";
				} else {
					return prefix + "[" + alias + "](" + alias + ")";
				}
			});
		});
	
		// Convert [Text](Alias) spans into links.
		var aliasRegexp = new RegExp(aliasMatch, "");
		this.converter.hooks.chain("preSpanGamut", function (text) {
			return text.replace(complexLinkRegexp, function (whole, prefix, text, alias) {
				if (alias.match(aliasRegexp)) {
					var url = "http://" + host + "/pages/" + alias;
					return prefix + "[" + text + "](" + url + ")";
				} else {
					return prefix + "[" + text + "](" + alias + ")";
				}
			});
		});

		if (pageId) {
			// Setup the editor stuff.
			this.editor = new Markdown.Editor(this.converter, pageId, {
				handler: function(){
					window.open("http://math.stackexchange.com/editing-help", "_blank");
				},
			});
		}
		
		InitMathjax(this.converter);

		if (this.editor) {
			this.editor.run();
		}
	};
});

// Directive for rendering markdown text.
app.directive("arbMarkdown", function ($compile, $timeout, pageService, markdownFactory) {
	return {
		template: "<div class='markdown-text'></div>",
		scope: {
			pageId: "@",
			useSummary: "@",
		},
		link: function(scope, element, attrs) {
			scope.page = pageService.pageMap[scope.pageId];
			var markdown = new markdownFactory();
			var host = window.location.host;

			// Convert page text to html.
			var html = markdown.converter.makeHtml(scope.useSummary ? scope.page.summary : scope.page.text);
			var $pageText = element.find(".markdown-text");
			$pageText.html(html);
			window.setTimeout(function() {
				MathJax.Hub.Queue(["Typeset", MathJax.Hub, $pageText.get(0)]);
			}, 100);
		
			// Setup attributes for page links that are within our domain.
			var re = new RegExp("^(?:https?:\/\/)?(?:www\.)?" + // match http and www stuff
				host + // match the url host part
				"\/pages\/" + aliasMatch + // [1] capture page alias
				"\/?" + // optional ending /
				"(.*)"); // optional other stuff
			$pageText.find("a").each(function(index, element) {
				var $element = $(element);
				var parts = $element.attr("href").match(re);
				if (parts === null) return;
				var pageAlias = parts[1];
	
				if ($element.hasClass("intrasite-link")) {
					return;
				}
				$element.addClass("intrasite-link").attr("page-id", pageAlias);
				// Check if we are embedding a vote
				if (parts[2].indexOf("embedVote") > 0) {
					$element.attr("embed-vote-id", pageAlias);
				} else if (pageAlias in pageService.pageMap) {
					if (pageService.pageMap[pageAlias].isDeleted()) {
						// Link to a deleted page.
						$element.addClass("red-link");
					} else {
						// Normal healthy link!
					}
				} else {
					// Mark as red link
					$element.attr("href", $element.attr("href").replace(/pages/, "edit"));
					$element.addClass("red-link");
				}
			});

			// Setup attributes for user links that are within our domain.
			var re = new RegExp("^(?:https?:\/\/)?(?:www\.)?" + // match http and www stuff
				host + // match the url host part
				"\/user\/" + aliasMatch + // [1] capture user alias
				"\/?" + // optional ending /
				"(.*)"); // optional other stuff
			$pageText.find("a").each(function(index, element) {
				var $element = $(element);
				var parts = $element.attr("href").match(re);
				if (parts === null) return;
				var userAlias = parts[1];
	
				if ($element.hasClass("user-link")) {
					return;
				}
				$element.addClass("user-link").attr("user-id", userAlias);

				// Do we want red links for invalid users?
/*
				if (userAlias in userService.userMap) {
					if (userService.userMap[userAlias].isDeleted()) {
						// Link to a deleted user.
						$element.addClass("red-link");
					} else {
						// Normal healthy link!
					}
				} else {
					// Mark as red link
					$element.addClass("red-link");
				}
*/
			});

		},
	};
});
