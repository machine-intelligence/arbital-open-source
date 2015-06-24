var zndMarkdown = zndMarkdown || function() {
	// Set up markdown editor and conversion.
	function init(inEditMode, pageId, pageText, $topParent, autocompleteService) {
		var host = window.location.host;
		var converter = Markdown.getSanitizingConverter();
		/*converter.hooks.chain("preSpanGamut", function (text) {
			console.log("text: " + text);
			return text.replace(/(.*?)"""(.*?)"""(.*)/g, "$1<u>$2</u>$3");
		});*/
	
		// Convert <summary> tags into a summary paragraph.
		converter.hooks.chain("preBlockGamut", function (text, rbg) {
			return text.replace(/^ {0,3}<summary> *\n(.+?)\n {0,3}<\/summary> *$/m, function (whole, inner) {
				var s = "\n\n**Summary:** " + inner + "\n\n";
				return rbg(s);
			});
		});
	
		// Convert <embed> tags into a link.
		var firstPass = !inEditMode;
		converter.hooks.chain("preBlockGamut", function (text, rbg) {
			return text.replace(/ {0,3}<embed> *(.+) *<\/embed> */g, function (whole, inner) {
				var s = "";
				if (firstPass) {
					s = "[LOADING](" + inner + "?embed=true)";
				} else {
					s = "[EMBEDDED PAGE](" + inner + ")";
				}
				return rbg(s);
			});
		});
	
		// Convert [[Text]]((Alias)) spans into links.
		var noBacktick = "(^|\\\\`|[^`])";
		var compexLinkRegexp = new RegExp(noBacktick + 
			"\\[\\[([^[\\]()]+?)\\]\\]" + // match [[Text]]
			"\\(\\(([A-Za-z0-9_-]+?)\\)\\)", "g"); // match ((Alias))
		converter.hooks.chain("preSpanGamut", function (text) {
			return text.replace(compexLinkRegexp,
				function (whole, prefix, text, alias) {
					var url = "http://" + host + "/pages/" + alias;
					return prefix + "[" + text + "](" + url + ")";
			});
		});
	
		// Convert [[Alias]] spans into links.
		var simpleLinkRegexp = new RegExp(noBacktick + 
				"\\[\\[([A-Za-z0-9_-]+?)\\]\\]", "g");
		converter.hooks.chain("preSpanGamut", function (text) {
			return text.replace(simpleLinkRegexp, function (whole, prefix, alias) {
				// TODO; do something other than ?customText=false, since that appears
				// in the URL if the user clicks on the link.
				var url = "http://" + host + "/pages/" + alias + "/?customText=false";
				var pageTitle = alias;
				if (autocompleteService && autocompleteService.aliasSource.length > 0){
					if (alias in autocompleteService.aliasMap) {
						pageTitle = autocompleteService.aliasMap[alias].PageTitle;
					}
				} else {
					if (alias in pageAliases) {
						pageTitle = pageAliases[alias].title;
					}
				}
				return prefix + "[" + pageTitle + "](" + url + ")";
			});
		});
	
		/*converter.hooks.chain("postNormalization", function (text, runSpanGamut) {
			return text.replace(/(.+?)( {0,2}\n)(.[^]*?\n)?([\n]{1,})/g, "$1[[[[1]]]]$2$3$4");
			//return text;
			//return text + "[[[[" + Math.floor(Math.random() * 1000000000) + "]]]]";
			/*return text.replace(/^ {0,3}""" *\n((?:.*?\n)+?) {0,3}""" *$/gm, function (whole, inner) {
				return "<blockquote>" + runBlockGamut(inner) + "</blockquote>\n";
			});
		});*/
	
		if (inEditMode) {
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
		InitMathjax(converter, undefined, pageId);
	
		var html = converter.makeHtml(pageText);
		var $pageText = $topParent.find(".page-text")
		$pageText.html(html);
		firstPass = false;
	
		// Setup attributes for links that are within our domain.
		var re = new RegExp("^(?:https?:\/\/)?(?:www\.)?" + // match http and www stuff
			host + // match the url host part
			"\/pages\/([A-Za-z0-9_-]+)" + // [1] capture page alias
			"(?:\/([0-9]+))?" + // [2] optionally capture privacyId
			"\/?" + // optional ending /
			"(\\?embed=true)?" + // [3] optionally capture embed param
			"(\\?customText=false)?"); // [4] optionally capture customText param
		function processLinks($div, fetchEmbeddedPages) {
			$div.find("a").each(function(index, element) {
				var $element = $(element);
				var parts = $element.attr("href").match(re);
				if (parts === null) return;
				if ($element.hasClass("intrasite-link")) {
					return;
				}
				$element.addClass("intrasite-link").attr("page-id", parts[1]).attr("privacy-key", parts[2]);
				if (gPageLinks !== undefined && gPageLinks[parts[1]] === false) {
					$element.attr("href", $element.attr("href").replace(/pages/, "edit"));
					$element.addClass("red-link");
				}
				if (parts[4] !== undefined) {
					$element.attr("href", $element.attr("href").replace("?customText=false", ""));
				}
				var $parent = $element.parent();
				var doEmbed = fetchEmbeddedPages && (parts[3] !== undefined);
				var data = {pageAlias: parts[1], privacyKey: parts[2], includeText: doEmbed};
				$.ajax({
					type: "POST",
					url: "/pageInfo/",
					data: JSON.stringify(data),
				})
				.success(function(r) {
					var page = JSON.parse(r);
					if (!doEmbed) {
						if (parts[4] !== undefined) {
							$element.text(page.Title);
						}
						return;
					}
					var $embeddedDiv = $("#embedded-page-template").clone().show()
					var $pageBody = $embeddedDiv.find(".embedded-page-body");
					var $title = $embeddedDiv.find(".embedded-page-title");
					$embeddedDiv.attr("id", "embedded-page" + page.PageId);
					$title.text(page.Title);
					$title.attr("href", "http://" + host + "/pages/" + page.PageId + "/" +
						(page.PrivacyKey > 0 ? page.PrivacyKey : ""));
					$embeddedDiv.find(".embedded-page-text").html(converter.makeHtml(page.Text));
					$parent.append($embeddedDiv);
					$element.remove();
					if (page.HasVote) {
						createVoteSlider($embeddedDiv.find(".embedded-vote-container"), page.PageId, page.Votes);
					}
					processLinks($embeddedDiv, false);
					setupIntrasiteLink($embeddedDiv.find(".intrasite-link"));
	
					// Set up toggle button
					$embeddedDiv.find(".hide-embedded-page").on("click", function(event) {
						var $target = $(event.target);
						$pageBody.slideToggle({});
						$target.toggleClass("glyphicon-triangle-bottom").toggleClass("glyphicon-triangle-right");
						return false;
					});
				});
			});
		};
		processLinks($pageText, true);
	};
	return {init: init};
}();
