"use strict";

// Used to escape regexp symbols in a string to make it safe for injecting into a regexp
RegExp.escape = function(s) {
    return s.replace(/[-\/\\^$*+?.()|[\]{}]/g, "\\$&");
};

// Keep the given div in a fixed position when the window is scrolled.
var keepDivFixed = function($div, offsetY) {
	// Make sure it's always in the top right corner.
	var divIsFixed = false;
	$div.css("left", $div.offset().left);
	var initialY = $div.offset().top + 20;
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

// Turn a callback function into a cleverly throttled version.
// Callback parameter should return true if the lock is to be set.
// Basically, we want:
// 1) Instant callback if the delay is met
// 2) Otherwise, wait to call the callback until delay is met
// 3) If we are waiting on the delay, don't stack additional calls
var createThrottledCallback = function(callback, delay) {
	// waitLock is set when we are waiting on the delay.
	var waitLock = false;
	// Timeout is set when we need to do the callback after the delay
	var timeout = undefined;

	var result = function() {
		if (waitLock) {
			if (!timeout) {
				timeout = window.setTimeout(function() {
					timeout = undefined;
					result();
				}, delay);
			}
			return;
		}
		if (callback()) {
			waitLock = true;
			window.setTimeout(function() {
				waitLock = false;
			}, delay);
		}
	};
	return result;
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
	options.showDelay = options.showDelay || 500;
	options.hideDelay = options.hideDelay || 200;

	// Create manually controlled popover.
	popoverOptions.trigger = "manual";
	$anchor.popover(popoverOptions);

	var firstTimeShow = true, isVisible = false, anchorHovering = false, popoverHovering = false;
	// Timeout for showing/hiding the popover.
	var timeout = undefined;
	// Hide the popover if the user is not hovering over anything.
	var hidePopover = function() {
		if (anchorHovering || popoverHovering || !isVisible) return;
		$anchor.popover("hide");
		if (options.uniqueName) {
			delete popoverMap[options.uniqueName];
		}
	};
	// Process when popover is going to be hidden.
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
			var $popover = $("#" + $anchor.attr("aria-describedby"));
			$popover.on("mouseenter", function(event){
				popoverHovering = true;
			});
			$popover.on("mouseleave", function(event){
				popoverHovering = false;
				timeout = setTimeout(hidePopover, options.hideDelay);
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

// serializeFormData takes input values from the given form and returns them as
// a map. Optionally, data can have pre-existing map values.
var serializeFormData = function($form, data) {
	if (data === undefined) data = {};
	$.each($form.serializeArray(), function(i, field) {
		data[field.name] = field.value;
	});
	data["__formSerialized"] = true;
	return data;
}

// submitForm handles the common functionality in submitting a form like
// showing/hiding UI elements and doing the AJAX call.
var submitForm = function($form, url, data, success, error) {
	var $errorText = $form.find(".submit-form-error");
	var $successText = $form.find(".submit-form-success");
	var invisibleSubmit = data["__invisibleSubmit"];
	if (!invisibleSubmit) {
		$form.find("[toggle-on-submit]").toggle();
	}

	if (!("__formSerialized" in data)) {
		serializeFormData($form, data);
	}

	console.log("Sending POST to " + url + ":"); console.log(data);
	$.ajax({
		type: "POST",
		url: url,
		data: JSON.stringify(data),
	})
	.always(function(r) {
		if (!invisibleSubmit) {
			$form.find("[toggle-on-submit]").toggle();
		}
	}).success(function(r) {
		$errorText.hide();
		$successText.show();
		success(r);
	}).fail(function(r) {
		// Want to show an error even on invisible submit.
		$errorText.show();
		$errorText.text(r.statusText + ": " + r.responseText);
		$successText.hide();
		console.log(r);
		if (error) error();
	});
}

