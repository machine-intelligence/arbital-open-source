// Keep the given div in a fixed position when the window is scrolled.
var keepDivFixed = function($div, offsetY) {
	// Make sure it's always in the top right corner.
	var divIsFixed = false;
	$div.css("left", $div.offset().left);
	var initialY = $div.offset().top;
	var qButtonOffsetY = 20;
	$(window).scroll(function(){
		var isFixed = $(window).scrollTop() > (initialY - qButtonOffsetY);
		if (isFixed !== divIsFixed) {
			if (!isFixed) {
				$div.css("position", "initial");
			} else {
				$div.css("position", "fixed").css("top", qButtonOffsetY);
			}
		}
		divIsFixed = isFixed;
	});
};

// Set up a popover attached to the given anchor. The popover will be displayed
// while the user is hovering over the anchor or the popover.
// options {
//   showDelay: how long (ms) to wait before showing popover
//   hideDelay: how long (ms) to wait to hide popover after the mouse leaves link & popover
//   uniqueName: if set, there will only be one popup visible with this name
// }
var popoverMap = {}; // uniqueName -> currently active popover's anchor
var createHoverablePopover = function($anchor, popoverOptions, options) {
	options = options || {};
	options.showDelay = options.showDelay || 300;
	options.hideDelay = options.hideDelay || 500;

	// Create manually controlled popover.
	popoverOptions.trigger = "manual";
	$anchor.popover(popoverOptions);

	var firstTimeShow = true, isVisible = false, anchorHovering = false, popoverHovering = false;
	var timeout = undefined;
	// Hide the popover if the user is not hovering over anything.
	var hidePopover = function() {
		if (anchorHovering || popoverHovering || !isVisible) return;
		$anchor.popover("hide");
		if (options.uniqueName) {
			delete popoverMap[options.uniqueName];
		}
	};
	$anchor.on("hide.bs.popover", function () {
		isVisible = false;
		if (timeout) clearTimeout(timeout);
	});

	var showPopover = function() {
		if (isVisible) return;
		$anchor.popover("show");
		if (options.uniqueName) {
			if (popoverMap[options.uniqueName]) {
				popoverMap[options.uniqueName].popover("hide");
			}
			popoverMap[options.uniqueName] = $anchor;
		}
		isVisible = true;

		if (firstTimeShow) {
			firstTimeShow = false;
			var $popover = $anchor.siblings(".popover");
			$popover.on("mouseenter", function(event){
				popoverHovering = true;
			});
			$popover.on("mouseleave", function(event){
				popoverHovering = false;
				setTimeout(hidePopover, options.hideDelay);
			});
		}
	};
	$anchor.on("mouseenter", function(event) {
		anchorHovering = true;
		if (timeout) clearTimeout(timeout);
		timeout = setTimeout(showPopover, options.showDelay);
	});
	$anchor.on("mouseleave", function(event) {
		anchorHovering = false;
		if (timeout) clearTimeout(timeout);
		timeout = setTimeout(hidePopover, options.hideDelay);
	});
	return $anchor;
};

// Create popover that tells the user they need to sign up / log in. The popover
// will be attached to the given $anchor element.
var showSignupPopover = function($anchor) {
	var options = {
		html: true,
		placement: "auto",
		title: "Login required",
		trigger: "hover",
		content: function() {
			var $content = $("<div>" + $("#signup-popover-template").html() + "</div>");
			return $content.html();
		}
	};
	$anchor.popover(options).popover("show");
	$anchor.on("hidden.bs.popover", function () {
		$anchor.popover("destroy");
	});
};

// Just a wrapper to get node's class name, but convert undefined into "".
var getNodeClassName = function(node) {
	return node.className || "";
}

// Return the parent node that's just under markdown-text and contains the given node.
var getParagraphNode = function(node) {
	var paragraphNode = node.parentNode;
	while (paragraphNode.parentNode != null) {
		if (getNodeClassName(paragraphNode.parentNode).indexOf("markdown-text") >= 0) {
			return paragraphNode;
		}
		paragraphNode = paragraphNode.parentNode;
	}
	return null;
}

// Called when the user selects markdown text.
// Return y position of where comment div should appear; null if it should
// be hidden.
var processSelectedParagraphText = function() {
	var selection = getStartEndSelection();
	if (!selection) return null;

	// Check that at least the start of the selection is within markdown-text.
	if (!getParagraphNode(selection.startContainer)) {
		return null;
	}
	var yOffset = $(selection.startContainer.parentElement).offset().top;
	if (getParagraphNode(selection.endContainer)) {
		// Middle between start and end.
		yOffset = (yOffset + $(selection.endContainer.parentElement).offset().top) / 2;
	}
	return yOffset;
};

// Wrap the given range in a a higlight node. That node gets the optinal nodeClass.
var highlightRange = function(range, nodeClass) {
	var parentNodeName = range.startContainer.parentNode.nodeName;
	if (parentNodeName === "SCRIPT") return;
	var startNodeName = range.startContainer.nodeName;
	var newNode = document.createElement((startNodeName == "DIV" || startNodeName == "P") ? "DIV" : "SPAN");
	if (nodeClass) {
		newNode.className += nodeClass;
	} else {
		newNode.className += "inline-comment-highlight";
	}
	newNode.appendChild(range.extractContents());
	range.insertNode(newNode);
};

// Return {context: paragraph text, text: selected text} object or null based
// on current user text selection.
var getSelectedParagraphText = function() {
	var selection = getStartEndSelection();
	if (!selection) return null;

	// Find the paragraph node, i.e. parent node right under markdown-text.
	var paragraphNode = getParagraphNode(selection.startContainer);
	if (!paragraphNode) return null;

	return getParagraphText(paragraphNode, selection);
}

// Return {context: paragraph text, text: selected text} object or null based
// on the given selection.
var getParagraphText = function(paragraphNode, selection) {
	var result = {text: "", context: "", offset: 0};
	// Whether the nodes we are visiting right now are inside the selection
	var insideText = false;
	// Store ranges we want to highlight.
	var ranges = [];
	// Compute context and text.
	recursivelyVisitChildren(paragraphNode, function(node, nodeText, needsEscaping) {
		var getEscapedText = function(start, end) {
			if (!needsEscaping) return nodeText.substring(start, end);
			return escapeMarkdownChars(nodeText.substring(start, end));
		};
		var escapedText;
		if (nodeText !== null) {
			escapedText = getEscapedText();
			result.context += escapedText;
		}
		if (!selection) return false;

		// If we are working with a selection, process that.
		var range = document.createRange();
		range.selectNodeContents(node);
		if (node == selection.startContainer && node == selection.endContainer) {
			if (nodeText !== null) {
				result.offset = result.context.length - nodeText.length + selection.startOffset;
				var offsetStr = nodeText.substring(0, selection.startOffset);
				if (needsEscaping) offsetStr = escapeMarkdownChars(offsetStr);
				result.offset = result.context.length - escapedText.length + offsetStr.length;
				result.text += getEscapedText(selection.startOffset, selection.endOffset);
				range.setStart(node, selection.startOffset);
				range.setEnd(node, selection.endOffset);
				ranges.push(range);
			} else {
				result.offset = result.context.length;
			}
		} else if (node == selection.startContainer) {
			insideText = true;
			if (nodeText !== null) {
				var offsetStr = nodeText.substring(0, selection.startOffset);
				if (needsEscaping) offsetStr = escapeMarkdownChars(offsetStr);
				result.offset = result.context.length - escapedText.length + offsetStr.length;
				result.text += getEscapedText(selection.startOffset);
				range.setStart(node, selection.startOffset);
				ranges.push(range);
			} else {
				result.offset = result.context.length;
			}
		} else if (node == selection.endContainer) {
			insideText = false;
			if (nodeText !== null) {
				result.text += getEscapedText(0, selection.endOffset);
				range.setEnd(node, selection.endOffset);
				ranges.push(range);
			}
		} else if(insideText) {
			if (nodeText !== null) {
				result.text += escapedText;
				ranges.push(range);
			}
		}
		return false;
	});
	// Highlight ranges after we did DOM traversal, so that there are no
	// modifications during the traversal.
	for (var i = 0; i < ranges.length; i++) {
		highlightRange(ranges[i]);
	}
	return result;
};

// Return true if the given node is a node containing MathJax spans.
var isNodeMathJax = function(node) {
	return node.className && node.className.indexOf("MathJax") >= 0;
};

// Recursively visit all leaf nodes in-order starting from the given node.
// Call callback for each visited node. Callback should return "true" iff the
// iteration is to be terminated.
// These are chars we need to escape when we examine nodes.
var	recursivelyVisitChildren = function(node, callback) {
	var done = false;
	var childLength = node.childNodes.length;
	var text = node.textContent;
	var needsEscaping = false;
	if (isNodeMathJax(node)) {
		text = "";
		childLength = 0;
	} else if (node.parentNode.id && node.parentNode.id.match(/^MathJax-Element-[0-9]+$/)) {
		childLength = 0;
		if (node.parentNode.type && node.parentNode.type.indexOf("mode=display") >= 0) {
			text = "$$$" + text + "$$$";
		} else {
			text = "$$" + text + "$$";
		}
	} else if (childLength === 0) {
		needsEscaping = true;
	}
	for (var n = 0; n < childLength; n++) {
		done = recursivelyVisitChildren(node.childNodes[n], callback);
		if (done) return done;
	}
	if (childLength === 0) {
		done = callback(node, text, needsEscaping);
	} else {
		done = callback(node, null);
	}
	return done;
}

// Return our type of Selection object.
var getStartEndSelection = function() {
	var selection = window.getSelection();
	if (selection.isCollapsed) return null;

	var r = document.createRange();
	var position = selection.anchorNode.compareDocumentPosition(selection.focusNode);
	if (position & Node.DOCUMENT_POSITION_PRECEDING) {
		// If text is selected right to left, swap the nodes.
		r.setStart(selection.focusNode, selection.focusOffset);
		r.setEnd(selection.anchorNode, selection.anchorOffset);
	} else {
		r.setStart(selection.anchorNode, selection.anchorOffset);
		r.setEnd(selection.focusNode, selection.focusOffset);
	}

	// If a node is inside Markdown block, we want to surround it.
	var parentNode = r.startContainer.parentNode;
	while (parentNode != null) {
		if (isNodeMathJax(parentNode)) {
			r.setStart(parentNode, 0);
		}
		parentNode = parentNode.parentNode;
	}

	// If the end not is inside the Mathjax block, it's a bit more complicatd.
	// We want to find the corresponding <script> element and surround up to it.
	parentNode = r.endContainer.parentNode;
	while (parentNode != null) {
		if (isNodeMathJax(parentNode)) {
			var match = parentNode.id.match(/(MathJax-Element-[0-9]+)-Frame/);
			if (match) {
				r.setEnd(document.getElementById(match[1]), 0); // <script>
			}
		}
		parentNode = parentNode.parentNode;
	}
	return r;
};

// Return the string, but with all the markdown chars escaped.
// NOTE: this doesn't work perfectly.
var escapeMarkdownChars = function(s) {
	return s.replace(/([\\`*_{}[\]()#+\-.!$])/g, "\\$1");
}
var unescapeMarkdownChars = function(s) {
	return s.replace(/\\([\\`*_{}[\]()#+\-.!$])/g, "$1");
}

// Send a new probability vote value to the server.
var postNewVote = function(pageId, value) {
	var data = {
		pageId: pageId,
		value: value,
	};
	$.ajax({
		type: "POST",
		url: "/newVote/",
		data: JSON.stringify(data),
	})
	.done(function(r) {
	});
}
	
// Set up a new vote slider. Set the slider's value based on the user's vote.
var createVoteSlider = function($parent, userService, page, isPopoverVote) {
	var userId = userService.user.id;
	// Convert votes into a user id -> {value, createdAt} map
	var voteMap = {};
	if (page.votes) {
		for(var i = 0; i < page.votes.length; i++) {
			var vote = page.votes[i];
			voteMap[vote.userId] = {value: vote.value, createdAt: vote.createdAt};
		}
	}

	// Copy vote-template and add it to the parent.
	var $voteDiv = $("#vote-template").clone().show().attr("id", "vote" + page.pageId).appendTo($parent);
	var $input = $voteDiv.find(".vote-slider-input");
	$input.attr("data-slider-id", $input.attr("data-slider-id") + page.pageId);
	var userVoteStr = userId in voteMap ? ("" + voteMap[userId].value) : "";
	var mySlider = $input.bootstrapSlider({
		step: 1,
		precision: 0,
		selection: "none",
		handle: "square",
		value: +userVoteStr,
		ticks: [1, 10, 20, 30, 40, 50, 60, 70, 80, 90, 99],
		formatter: function(s) { return s + "%"; },
	});
	var $tooltip = $parent.find(".tooltip-main");

	// Set the value of the user's vote.
	var setMyVoteValue = function($voteDiv, userVoteStr) {
		$voteDiv.attr("my-vote", userVoteStr);
		$voteDiv.find(".my-vote").toggle(userVoteStr !== "");
		$voteDiv.find(".my-vote-value").text("| my vote is \"" + (+userVoteStr) + "%\"");
	}
	setMyVoteValue($voteDiv, userVoteStr);

	// Setup vote bars.
	// A bar represents users' votes for a given value. The tiled background
	// allows us to display each vote separately.
	var bars = {}; // voteValue -> {bar: jquery bar element, users: array of user ids who voted on this value}
	// Stuff for setting up the bars' css.
	var $barBackground = $parent.find(".bar-background");
	var $sliderTrack = $parent.find(".slider-track");
	var originLeft = $sliderTrack.offset().left;
	var originTop = $sliderTrack.offset().top;
	var barWidth = Math.max(5, $sliderTrack.width() / (99 - 1) * 2);
	// Set the correct css for the given bar object given the number of votes it has.
	var setBarCss = function(bar) {
		var $bar = bar.bar;
		var voteCount = bar.users.length;
		$bar.toggle(voteCount > 0);
		$bar.css("height", 11 * voteCount);
		$bar.css("z-index", 2 + voteCount);
		$barBackground.css("height", Math.max($barBackground.height(), $bar.height()));
		$barBackground.css("top", 10);
	}
	var highlightBar = function($bar, highlight) {
		var css = "url(/static/images/vote-bar.png)";
		var highlightColor = "rgba(128, 128, 255, 0.3)";
		if(highlight) {
			css = "linear-gradient(" + highlightColor + "," + highlightColor + ")," + css;
		}
		$bar.css("background", css);
		$bar.css("background-size", "100% 11px"); // have to set this each time
	};
	// Get the bar object corresponding to the given vote value. Create a new one if there isn't one.
	var getBar = function(vote) {
		if (!(vote in bars)) {
			var x = (vote - 1) / (99 - 1);
			var $bar = $("<div class='vote-bar'></div>");
			$bar.css("left", x * $sliderTrack.width() - barWidth / 2);
			$bar.css("width", barWidth);
			$barBackground.append($bar);
			bars[vote] = {bar: $bar, users: []};
		}
		return bars[vote];
	}
	for(var id in voteMap){
		// Create stacks for all the votes.
		var bar = getBar(voteMap[id].value);
		bar.users.push(id);
		setBarCss(bar);
	}

	// Convert mouse X position into % vote value.
	var voteValueFromMousePosX = function(mouseX) {
		var x = (mouseX - $sliderTrack.offset().left) / $sliderTrack.width();
		x = Math.max(0, Math.min(1, x));
		return Math.round(x * (99 - 1) + 1);
	};

	// Update the label that shows how many votes have been cast.
	var updateVoteCount = function() {
		var votesLength = Object.keys(voteMap).length;
		$voteDiv.find(".vote-count").text(votesLength + " vote" + (votesLength == 1 ? "" : "s"));
	};
	updateVoteCount();

	// Set handle's width.
	var $handle = $parent.find(".min-slider-handle");
	$handle.css("width", barWidth);

	// Don't track mouse movements and such for the vote in a popover.
	if (isPopoverVote) {
		if (!(userId in voteMap)) {
			$handle.hide();
		}
		return;
	}

	// Catch mousemove event on the text, so that it doesn't propagate to parent
	// and spawn popovers, which will prevent us clicking on "x" button to delete
	// our vote.
	$parent.find(".text-center").on("mousemove", function(event){
		return false;
	});

	var mouseInParent = false;
	var mouseInPopover = false;
	// Destroy the popover that displayes who voted for a given value.
	var $usersPopover = undefined;
	var destroyUsersPopover = function() {
		if ($usersPopover !== undefined) {
			$usersPopover.popover("destroy");
			highlightBar($usersPopover, false);
		}
		$usersPopover = undefined;
		mouseInPopover = false;
	};

	// Track mouse movement to show voter names.
	$parent.on("mouseenter", function(event) {
		mouseInParent = true;
		$handle.show();
		$tooltip.css("opacity", 1.0);
	});
	$parent.on("mouseleave", function(event) {
		mouseInParent = false;
		if (!(userId in voteMap)) {
			$handle.hide();
		} else {
			$input.bootstrapSlider("setValue", voteMap[userId].value);
		}
		$tooltip.css("opacity", 0.0);
		if (!mouseInPopover) {
			destroyUsersPopover();
		}
	});
	$parent.trigger("mouseleave");
	$parent.on("mousemove", function(event) {
		// Update slider.
		var voteValue = voteValueFromMousePosX(event.pageX);
		$input.bootstrapSlider("setValue", voteValue);
		if (mouseInPopover) return true;

		// We do a funky search to see if there is a vote nearby, and if so, show popover.
		var offset = 0, maxOffset = 5;
		var offsetSign = 1;
		var foundBar = false;
		while(offset <= maxOffset) {
			if(offsetSign < 0) offset++;
			offsetSign = -offsetSign;
			var hoverVoteValue = voteValue + offsetSign * offset;
			if (!(hoverVoteValue in bars)) {
				continue;
			}
			foundBar = true;
			var bar = bars[hoverVoteValue];
			// Don't do anything if it's still the same bar as last time.
			if (bar.bar === $usersPopover) {
				break;
			}
			// Destroy old one.
			destroyUsersPopover();
			// Create new popover.
			$usersPopover = bar.bar;
			highlightBar(bar.bar, true);
			$usersPopover.popover({
				html : true,
				placement: "bottom",
				trigger: "manual",
				title: "Voters (" + hoverVoteValue + "%)",
				content: function() {
					var html = "";
					for(var i = 0; i < bar.users.length; i++) {
						var userId = bar.users[i];
						var user = userService.userMap[userId];
						var name = user.firstName + "&nbsp;" + user.lastName;
						html += "<a href='" + userService.getUserUrl(userId) + "'>" + name + "</a> " +
							"<span class='gray-text'>(" + voteMap[userId].createdAt + ")</span><br>";
					}
					return html;
				}
			}).popover("show");
			var $popover = $barBackground.find(".popover");
			$popover.on("mouseenter", function(event){
				mouseInPopover = true;
			});
			$popover.on("mouseleave", function(event){
				mouseInPopover = false;
				if (!mouseInParent) {
					destroyUsersPopover();
				}
			});
			break;
		}
		if (!foundBar) {
			// We didn't find a bar, so destroy any existing popover.
			destroyUsersPopover();
		}
	});

	// Handle user's request to delete their vote.
	$voteDiv.find(".delete-my-vote-link").on("click", function(event) {
		var bar = bars[voteMap[userId].value];
		bar.users.splice(bar.users.indexOf(userId), 1);
		setBarCss(bar);
		if (bar.users.length <= 0){
			delete bars[voteMap[userId].value];
		}

		mouseInPopover = false;
		mouseInParent = false;
		delete voteMap[userId];
		$parent.trigger("mouseleave");
		$parent.trigger("mouseenter");

		postNewVote(page.pageId, 0.0);
		setMyVoteValue($voteDiv, "");
		updateVoteCount();
		return false;
	});
	
	// Track click to see when the user wants to vote / update their vote.
	$parent.on("click", function(event) {
		if (mouseInPopover) return true;
		if (userId === "0") {
			showSignupPopover($(event.currentTarget));
			return true;
		}
		if (userId in voteMap && voteMap[userId].value in bars) {
			// Update old bar.
			var bar = bars[voteMap[userId].value];
			bar.users.splice(bar.users.indexOf(userId), 1);
			setBarCss(bar);
			destroyUsersPopover();
			if (bar.users.length <= 0) {
				delete bars[voteMap[userId].value];
			}
		}

		// Set new vote and update all the things.
		var vote = voteValueFromMousePosX(event.pageX); 
		voteMap[userId] = {value: vote, createdAt: "now"};
		postNewVote(page.pageId, vote);
		setMyVoteValue($voteDiv, "" + vote);
		updateVoteCount();

		// Update new bar.
		var bar = getBar(vote);
		bar.users.push(userId);
		setBarCss(bar);
	});
}
