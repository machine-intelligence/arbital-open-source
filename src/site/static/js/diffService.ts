'use strict';
// jscs:disable requireCamelCaseOrUpperCaseIdentifiers

// Service for creating diffs.
app.service('diffService', function() {
	var that = this;

	// Return the title + clickbait + text from the given edit, ready to diff
	this.getTextToDiff = function(edit) {
		return 'Title: ' + edit.title + '\n' + 'Clickbait: ' + edit.clickbait + '\n\n' + edit.text;
	};

	// Return HTML corresponding to the diff between two texts.
	this.getDiffHtml = function(thisEdit, thatEdit, opt_expandDiffs) {
		var thisText = that.getTextToDiff(thisEdit);
		var thatText = that.getTextToDiff(thatEdit);
		var dmp = new diff_match_patch(); // jscs:ignore requireCapitalizedConstructors
		var diffs = dmp.diff_main(thisText, thatText);
		dmp.diff_cleanupSemantic(diffs);

		if (!opt_expandDiffs) {
			diffs = this._getCollapsedDiffs(diffs);
		}

		var diffHtml = dmp.diff_prettyHtml(diffs).replace(/&para;/g, '');
		return diffHtml;
	};

	// Replace unchanged paragraphs with "..."
	this._getCollapsedDiffs = function(diffs) {
		return diffs.map(function(diff, index) {
			// Ignore diffs that contain changes
			if (diff[0] != 0) {
				return diff;
			}

			var diffString = diff[1];
			var breakBreakString = '\n\n';

			// Ignore diffs with only one instance of '\n\n'
			if (diffString.split(breakBreakString).length < 3) {
				return diff;
			}

			// Begin collapsing at the first instance of '\n\n', unless this is the first diff
			var beginCollapse = diffString.indexOf(breakBreakString) + breakBreakString.length;
			if (index == 0) {
				beginCollapse = 0;
			}

			// End collapsing at the last instance of '\n\n', unless this is the last diff
			var endCollapse = diffString.lastIndexOf(breakBreakString);
			if (index == diffs.length - 1) {
				endCollapse = diffString.length;
			}

			// Keep the bit before the first '↵↵' and the last '↵↵'. Replace the rest with '...'
			diffString = diffString.substring(0, beginCollapse) + '...' +
				diffString.substring(endCollapse);

			return [0, diffString];
		});
	};
});
