// Set up markdown editor and conversion.
function setUpMarkdown(inEditMode) {
	// TODO: get pageText and add all the comment tags
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

	// Convert [[Alias]] spans into links.
	converter.hooks.chain("preSpanGamut", function (text) {
		return text.replace(/\[\[([A-Za-z0-9_-]+?)\]\]/, function (whole, inner) {
			var url = "http://" + host + "/pages/" + inner;
			var pageTitle = inner;
			if (inner in pageAliases) {
				pageTitle = pageAliases[inner].title;
			}
			return "[" + pageTitle + "](" + url + ")";
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
		var editor = new Markdown.Editor(converter, "", {handler: function(){
			window.open("http://math.stackexchange.com/editing-help", "_blank");
		}});
		InitMathjax(converter, editor, "");
		editor.run();
		return;
	}
	InitMathjax(converter, undefined, "");

	var html = converter.makeHtml(gPageText);
	var $pageText = $(".page-text")
	$pageText.html(html);
	firstPass = false;

	// Setup attributes for links that are within our domain.
	var re = new RegExp("^(?:https?:\/\/)?(?:www\.)?" + // match http and www stuff
		host + // match the url host part
		"\/pages\/([A-Za-z0-9_-]+)" + // capture page alias
		"(?:\/([0-9]+))?" + // optionally capture privacyId
		"(?:\\?embed\=(true))?"); // optionally capture embed param
	function processLinks($div, fetchEmbeddedPages) {
		$div.find("a").each(function(index, element) {
			var $element = $(element);
			var parts = $element.attr("href").match(re);
			if (parts === null) return;
			if (!$element.hasClass("intrasite-link")) {
				$element.addClass("intrasite-link").attr("page-id", parts[1]).attr("privacy-key", parts[2]);
				if (parts[3] && fetchEmbeddedPages && !inEditMode) {
					var $parent = $element.parent();
					var data = {pageAlias: parts[1], privacyKey: parts[2], includeText: true};
					$.ajax({
						type: "POST",
						url: "/pageInfo/",
						data: JSON.stringify(data),
					})
					.success(function(r) {
						var page = JSON.parse(r);
						var $embeddedDiv = $("#embedded-page-template").clone().show().attr("id", "embedded-page" + page.PageId);
						var $pageBody = $embeddedDiv.find(".embedded-page-body");
						$embeddedDiv.find(".embedded-page-title").text(page.Title).attr("href", "http://" + host + "/pages/" + page.PageId + "/" + (page.PrivacyKey > 0 ? page.PrivacyKey : ""));
						$embeddedDiv.find(".embedded-page-text").html(converter.makeHtml(page.Text));
						$parent.append($embeddedDiv);
						$element.remove();
						if (page.HasVote) {
							createVoteSlider($embeddedDiv.find(".embedded-vote-container"), page.PageId, page.VoteCount,
								page.VoteValue.Valid ? "" + page.VoteValue.Float64 : "",
								page.MyVoteValue.Valid ? "" + page.MyVoteValue.Float64 : "");
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
				}
			}
		});
	}
	processLinks($pageText, true);
};
