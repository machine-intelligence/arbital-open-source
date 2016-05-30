// jscs:disable
'use strict';

var InitMathjax = (function() {
	var ready   = false;  // true after initial typeset is complete
	var pending = false;  // true when MathJax has been requested
	var preview = null;   // the preview container
	var inlineDelim = '$$';  // the inline math delimiter
	var blockDelim = '$$$';  // the block math delimeter

	var blocks, start, end, last, braces; // used in searching for math
	var math;                             // stores math until markdone is done

	var HUB = MathJax.Hub;

	//
	//  Runs after initial typeset
	//
	HUB.Queue(function() {
		ready = true;
		HUB.processUpdateTime = 10;  // reduce update time so that we can cancel easier
	});

	//
	//  The pattern for math delimiters and special symbols
	//    needed for searching for math in the page.
	//
	var SPLIT = /(\$\$?|\\(?:begin|end)\{[a-z]*\*?\}|\\[\\{}$]|[{}]|(?:\n\s*)+|@@\d+@@)/i;

	//
	//  The math is in blocks i through j, so
	//    collect it into one block and clear the others.
	//  Replace &, <, and > by named entities.
	//  For IE, put <br> at the ends of comments since IE removes \n.
	//  Clear the current math positions and store the index of the
	//    math, then push the math string onto the storage array.
	//
	function processMath(i, j) {
		var block = blocks.slice(i,j + 1).join('')
		.replace(/&/g,'&amp;')                   // use HTML entity for &
		.replace(/</g,'&lt;')                    // use HTML entity for <
		.replace(/>/g,'&gt;')                    // use HTML entity for >
		;
		if (HUB.Browser.isMSIE) {block = block.replace(/(%[^\n]*)\n/g,'$1<br/>\n');}
		while (j > i) {blocks[j] = ''; j--;}
		blocks[i] = '@@' + math.length + '@@'; math.push(block);
		start = end = last = null;
	}

	//
	//  Break up the text into its component parts and search
	//    through them for math delimiters, braces, linebreaks, etc.
	//  Math delimiters must match and braces must balance.
	//  Don't allow math to pass through a double linebreak
	//    (which will be a paragraph).
	//
	function removeMath(text) {
		start = end = last = null;       // for tracking math delimiters
		math = [];                       // stores math strings for latter

		blocks = text.replace(/\r\n?/g,'\n').split(SPLIT);
		for (var i = 1, m = blocks.length; i < m; i += 2) {
			var block = blocks[i];
			if (block.charAt(0) === '@') {
				//
				//  Things that look like our math markers will get
				//  stored and then retrieved along with the math.
				//
				blocks[i] = '@@' + math.length + '@@';
				math.push(block);
			} else if (start) {
				//
				//  If we are in math, look for the end delimiter,
				//    but don't go past double line breaks, and
				//    and balance braces within the math.
				//
				if (block === end) {
					if (braces) {last = i;} else {processMath(start,i);}
				} else if (block.match(/\n.*\n/)) {
					if (last) {i = last; processMath(start,i);}
					start = end = last = null; braces = 0;
				} else if (block === '{') {braces++;} else if (block === '}' && braces) {braces--;}
			} else {
				//
				//  Look for math start delimiters and when
				//    found, set up the end delimiter.
				//
				if (block === inlineDelim || block === blockDelim) {
					start = i; end = block; braces = 0;
				} else if (block.substr(1,5) === 'begin') {
					start = i; end = '\\end' + block.substr(6); braces = 0;
				}
			}
		}
		if (last) {processMath(start,last);}
		return blocks.join('');
	}

	//
	//  Put back the math strings that were saved,
	//    and clear the math array (no need to keep it around).
	//
	function replaceMath(text) {
		text = text.replace(/@@(\d+)@@/g,function(match, n) {return math[n];});
		math = null;
		return text;
	}

	//
	//  This is run to restart MathJax after it has finished
	//    the previous run (that may have been canceled)
	//
	function RestartMJ() {
		pending = false;
		HUB.cancelTypeset = false;   // won't need to do this in the future
		//HUB.Queue(['Typeset',HUB,preview]);
	}

	//
	//  When the preview changes, cancel MathJax and restart,
	//    if we haven't done that already.
	//
	function UpdateMJ() {
		if (!pending && ready) {
			pending = true;
			HUB.Cancel();
			HUB.Queue(RestartMJ);
		}
	}

	//
	//  Save the preview ID and the inline math delimiter.
	//  Create a converter for the editor and register a preConversion hook
	//   to handle escaping the math.
	//  Create a preview refresh hook to handle starting MathJax.
	//
	function prepareWmdForMathJax(converterObject, editorObject, wmdId) {
		if (!editorObject) {
			converterObject.hooks.chain('preConversion',removeMath);
			converterObject.hooks.chain('postConversion',replaceMath);
		}
		/*if (editorObject) {
			editorObject.hooks.chain('onPreviewRefresh', UpdateMJ);
			preview = document.getElementById('wmd-preview' + wmdId);
		}*/
		HUB.Config({
			// reduce chunk for more frequent updates
			'HTML-CSS': {EqnChunk: 10, EqnChunkFactor: 1, linebreaks: {automatic: true}}, 
			SVG: {EqnChunk: 10, EqnChunkFactor: 1, linebreaks: {automatic: true}},
			tex2jax: {
				inlineMath: [[inlineDelim, inlineDelim]],
				displayMath: [[blockDelim, blockDelim]],
				processEscapes: true,
			},
			showProcessingMessages: false,
		});
	}

	return prepareWmdForMathJax;
})();

//
//  Set up MathJax to allow canceling of typesetting, if it
//    doesn't already have that.
//
(function() {
	var HUB = MathJax.Hub;

	if (!HUB.Cancel) {

		HUB.cancelTypeset = false;
		var CANCELMESSAGE = 'MathJax Canceled';

		HUB.Register.StartupHook('HTML-CSS Jax Config',function() {
			var HTMLCSS = MathJax.OutputJax['HTML-CSS'], TRANSLATE = HTMLCSS.Translate;
			HTMLCSS.Augment({
				Translate: function(script, state) {
					if (HUB.cancelTypeset || state.cancelled) {throw Error(CANCELMESSAGE);}
					return TRANSLATE.call(HTMLCSS,script,state);
				}
			});
		});

		HUB.Register.StartupHook('SVG Jax Config',function() {
			var SVG = MathJax.OutputJax['SVG'], TRANSLATE = SVG.Translate;
			SVG.Augment({
				Translate: function(script, state) {
					if (HUB.cancelTypeset || state.cancelled) {throw Error(CANCELMESSAGE);}
					return TRANSLATE.call(SVG,script,state);
				}
			});
		});

		HUB.Register.StartupHook('TeX Jax Config',function() {
			var TEX = MathJax.InputJax.TeX, TRANSLATE = TEX.Translate;
			TEX.Augment({
				Translate: function(script, state) {
					if (HUB.cancelTypeset || state.cancelled) {throw Error(CANCELMESSAGE);}
					return TRANSLATE.call(TEX,script,state);
				}
			});
		});

		var PROCESSERROR = HUB.processError;
		HUB.processError = function(error, state, type) {
			if (error.message !== CANCELMESSAGE) {return PROCESSERROR.call(HUB,error,state,type);}
			MathJax.Message.Clear(0,0);
			state.jaxIDs = []; state.jax = {}; state.scripts = [];
			state.i = state.j = 0;
			state.cancelled = true;
			return null;
		};

		HUB.Cancel = function() {this.cancelTypeset = true;};
	}
})();
