// Just a wrapper to get node's class name, but convert undefined into "".
var getNodeClassName = function(node) {
	return node.className || '';
};

// Return the parent node that's just under markdown-text and contains the given node.
var getParagraphNode = function(node) {
	var paragraphNode = node.parentNode;
	while (paragraphNode.parentNode != null) {
		if (getNodeClassName(paragraphNode.parentNode).indexOf('markdown-text') >= 0) {
			return paragraphNode;
		}
		paragraphNode = paragraphNode.parentNode;
	}
	return null;
};

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
	var yOffset = $(selection.startContainer.parentNode).offset().top;
	if (getParagraphNode(selection.endContainer)) {
		// Middle between start and end.
		yOffset = (yOffset + $(selection.endContainer.parentNode).offset().top) / 2;
	}
	return yOffset;
};

// Wrap the given range in a a higlight node. That node gets the optinal nodeClass.
var highlightRange = function(range, nodeClass) {
	var parentNodeName = range.startContainer.parentNode.nodeName;
	if (parentNodeName === 'SCRIPT') return;
	var startNodeName = range.startContainer.nodeName;
	var newNode = document.createElement((startNodeName == 'DIV' || startNodeName == 'P') ? 'DIV' : 'SPAN');
	if (nodeClass) {
		newNode.className += nodeClass;
	} else {
		newNode.className += 'inline-comment-highlight';
	}
	newNode.appendChild(range.extractContents());
	range.insertNode(newNode);
};

// Return {context: paragraph text, text: selected text} object or null based
// on current user text selection.
// cachedSelection - if set, will use this selection instead of the current one
var getSelectedParagraphText = function(cachedSelection, skipHighlight) {
	var selection = cachedSelection || getStartEndSelection();
	if (!selection) return null;

	// Find the paragraph node, i.e. parent node right under markdown-text.
	var paragraphNode = getParagraphNode(selection.startContainer);
	if (!paragraphNode) return null;

	return getParagraphText(paragraphNode, selection, skipHighlight);
};

// Return {context: paragraph text, text: selected text} object or null based
// on the given selection.
var getParagraphText = function(paragraphNode, selection, skipHighlight) {
	var result = {text: '', context: '', offset: 0, paragraphNode: paragraphNode};
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
		} else if (insideText) {
			if (nodeText !== null) {
				result.text += escapedText;
				ranges.push(range);
			}
		}
		return false;
	});
	// Highlight ranges after we did DOM traversal, so that there are no
	// modifications during the traversal.
	if (!skipHighlight) {
		for (var i = 0; i < ranges.length; i++) {
			highlightRange(ranges[i]);
		}
	}
	return result;
};

// Return true if the given node is a node containing MathJax spans.
var isNodeMathJax = function(node) {
	return node.className && node.className.indexOf('MathJax') >= 0;
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
		text = '';
		childLength = 0;
	} else if (node.parentNode.id && node.parentNode.id.match(/^MathJax-Element-[0-9]+$/)) {
		childLength = 0;
		if (node.parentNode.type && node.parentNode.type.indexOf('mode=display') >= 0) {
			text = '$$$' + text + '$$$';
		} else {
			text = '$$' + text + '$$';
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
};

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
	return s.replace(/([\\`*_{}[\]()#+\-.!$])/g, '\\$1');
};
var unescapeMarkdownChars = function(s) {
	return s.replace(/\\([\\`*_{}[\]()#+\-.!$])/g, '$1');
};

// Inline comments
// Create the inline comment highlight spans for the given paragraph.
var createInlineCommentHighlight = function(paragraphNode, start, end, nodeClass) {
	// How many characters we passed in the anchor context (which has escaped characters).
	var charCount = 0;
	// Store ranges we want to highlight.
	var ranges = [];
	// Compute context and text.
	recursivelyVisitChildren(paragraphNode, function(node, nodeText, needsEscaping) {
		if (nodeText === null) return false;
		var escapedText = needsEscaping ? escapeMarkdownChars(nodeText) : nodeText;
		var nodeWholeTextLength = node.wholeText ? node.wholeText.length : 0;
		var range = document.createRange();
		var nextCharCount = charCount + escapedText.length;
		if (charCount <= start && nextCharCount >= end) {
			var pStart = unescapeMarkdownChars(escapedText.substring(0, start - charCount)).length;
			var pEnd = unescapeMarkdownChars(escapedText.substring(0, end - charCount)).length;
			range.setStart(node, pStart);
			range.setEnd(node, Math.min(nodeWholeTextLength, pEnd));
			ranges.push(range);
		} else if (charCount <= start && nextCharCount > start) {
			var pStart = unescapeMarkdownChars(escapedText.substring(0, start - charCount)).length;
			range.setStart(node, pStart);
			range.setEnd(node, Math.min(nodeWholeTextLength, nodeText.length));
			ranges.push(range);
		} else if (start < charCount && nextCharCount >= end) {
			range.setStart(node, 0);
			range.setEnd(node, Math.min(nodeWholeTextLength, end - charCount));
			ranges.push(range);
		} else if (start < charCount && nextCharCount > start) {
			if (nodeWholeTextLength > 0) {
				range.setStart(node, 0);
				range.setEnd(node, Math.min(nodeWholeTextLength, nodeText.length));
			} else {
				range.selectNodeContents(node);
			}
			ranges.push(range);
		} else if (start == charCount && charCount == nextCharCount) {
			// Rare occurence, but this captures MathJax divs/spans that
			// precede the script node where we actually get the text from.
			range.selectNodeContents(node);
			ranges.push(range);
		}
		charCount = nextCharCount;
		return charCount >= end;
	});
	// Highlight ranges after we did DOM traversal, so that there are no
	// modifications during the traversal.
	for (var i = 0; i < ranges.length; i++) {
		highlightRange(ranges[i], nodeClass);
	}
	return ranges.length > 0 ? ranges[0].startContainer : null;
};
