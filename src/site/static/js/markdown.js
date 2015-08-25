var zndMarkdown = zndMarkdown || function() {
	// Set up markdown editor and conversion.
	function init(inEditMode, pageId, pageText, $topParent, pageService, autocompleteService) {
		var page = pageService.pageMap[pageId];
		var host = window.location.host;
		var converter = Markdown.getSanitizingConverter();
	
		// Convert [Text](Alias) spans into links.
		var noBacktickOrBracket = "(^|\\\\`|\\\\\\[|[^`[])";
		var compexLinkRegexp = new RegExp(noBacktickOrBracket + 
			"\\[([^[\\]()]+?)\\]" + // match [Text]
			"\\(([A-Za-z0-9_-]+?)\\)", "g"); // match (Alias)
		var aliasRegexp = new RegExp("[A-Za-z0-9_-]+", "");
		converter.hooks.chain("preSpanGamut", function (text) {
			return text.replace(compexLinkRegexp, function (whole, prefix, text, alias) {
				if (alias.match(aliasRegexp)) {
					var url = "http://" + host + "/pages/" + alias;
					return prefix + "[" + text + "](" + url + ")";
				} else {
					return prefix + "[" + text + "](" + alias + ")";
				}
			});
		});

		// Convert [Alias] spans into links.
		var noParen = "(?=$|[^(])";
		var simpleLinkRegexp = new RegExp(noBacktickOrBracket + 
				"\\[([A-Za-z0-9_-]+?)\\]" + noParen, "g");
		converter.hooks.chain("preSpanGamut", function (text) {
			return text.replace(simpleLinkRegexp, function (whole, prefix, alias) {
				if (alias.match(aliasRegexp)) {
					var url = "http://" + host + "/pages/" + alias + "/";
					var pageTitle = alias;
					if (page.links && page.links[alias]) {
						pageTitle = page.links[alias];
					}
					return prefix + "[" + pageTitle + "](" + url + ")";
				} else {
					return prefix + "[" + text + "](" + alias + ")";
				}
			});
		});
	
		if (inEditMode) {
			// Setup the editor stuff.
			var editor = new Markdown.Editor(converter, pageId, {
				autocompleteService: autocompleteService,
				handler: function(){
					window.open("http://math.stackexchange.com/editing-help", "_blank");
				},
			});
			InitMathjax(converter, editor, pageId);
			editor.run();
			return;
		}
		InitMathjax(converter);
	
		// Convert page text to html.
		var html = converter.makeHtml(pageText);
		var $pageText = $topParent.find(".markdown-text")
		$pageText.html(html);
		window.setTimeout(function() {
			MathJax.Hub.Queue(["Typeset", MathJax.Hub, $pageText.get(0)]);
		}, 100);
	
		// Setup attributes for links that are within our domain.
		var re = new RegExp("^(?:https?:\/\/)?(?:www\.)?" + // match http and www stuff
			host + // match the url host part
			"\/pages\/([A-Za-z0-9_-]+)" + // [1] capture page alias
			"(?:\/([0-9]+))?" + // [2] optionally capture privacyId
			"\/?"); // optional ending /
		$pageText.find("a").each(function(index, element) {
			var $element = $(element);
			var parts = $element.attr("href").match(re);
			if (parts === null) return;
			if ($element.hasClass("intrasite-link")) {
				return;
			}
			$element.addClass("intrasite-link").attr("page-id", parts[1]).attr("privacy-key", parts[2]);
			if (page.links && page.links[parts[1]]) {
				// Good link!
			} else {
				$element.attr("href", $element.attr("href").replace(/pages/, "edit"));
				$element.addClass("red-link");
			}
		});
	};
	return {init: init};
}();
