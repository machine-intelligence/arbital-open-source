// Reload the page with a lastVisit parameter so we can pretend that we are
// looking at a page at that time. This way new/updated markers are displayed
// correctly.
function smartPageReload() {
	var url = $("body").attr("claim-url");
	var lastVisit = encodeURIComponent($("body").attr("last-visit"));
	window.location.replace(url + "?lastVisit=" + lastVisit);
}

// Replace Markdown text with corresponding HTML.
$(function() {
	var converter = Markdown.getSanitizingConverter();
	$(".claim-text").each(function(index, element) {
		//document.write(converter.makeHtml("**I am bold!**"));
		$(element).html(converter.makeHtml($(element).text()));
		var editor = new Markdown.Editor(converter, "-" + $(element).closest(".claim").attr("claim-id"));
		editor.run();
	});
});

$(function() {
	// Claim editing stuff.
	var toggleEditClaim = function($claim) {
		$claim.find(".claim-body").toggle();
		$claim.find(".edit-claim-form").toggle();
		$claim.find(".edit-claim-link").toggleClass("on");
	}
	$(".edit-claim-link").on("click", function(event) {
		var $target = $(event.target);
		var $claim = $target.closest(".claim");
		toggleEditClaim($claim);
		$claim.find(".edit-claim-summary").focus();
		return false;
	});
	$(".edit-claim-form").on("submit", function(event) {
		var $form = $(event.target);
		//var $claim = $form.closest(".claim");

		var data = {};
		submitForm($form, "/updateClaim/", data, function(r) {
			smartPageReload();
			/*var $claimUrl = $claim.find(".claimUrl");
			var url = $claim.find(".editClaimUrl").val();
			toggleEditClaim($claim);
			$claimUrl.attr("href", url);
			$claimUrl.toggle(url.length > 0);
			$claim.find(".claimText").text($claim.find(".editClaimTextarea").val());*/
		});
		return false;
	});

	// Deleting an input.
	$(".delete-input-link").on("click", function(event) {
		var $target = $(event.target);
		var data = {
			parentClaimId: $("body").attr("parent-claim-id"),
			childClaimId: $target.closest(".claim").attr("claim-id"),
		};
		$.ajax({
			type: 'POST',
			url: '/deleteInput/',
			data: JSON.stringify(data),
		})
		.done(function(r) {
			smartPageReload();
		});
	});

	// Comment editing stuff.
	function toggleEditComment($comment) {
		$comment.find(".comment-body").toggle();
		$comment.find(".edit-comment-form").toggle();
	}
	$(".edit-comment-link").on("click", function(event) {
		var $comment = $(event.target).closest(".comment-row").find(".comment");
		var $editCommentTextarea = $comment.find(".edit-comment-text");
		toggleEditComment($comment);
		$editCommentTextarea.focus();
		return false;
	});
	$(".edit-comment-form").on("submit", function(event) {
		var $form = $(event.target);
		var $comment = $form.closest(".comment");
		var $editCommentTextarea = $form.find(".edit-comment-text");
		var $commentText = $comment.find(".comment-text");

		var data = {id: $comment.attr("comment-id")};
		submitForm($form, "/updateComment/", data, function(r) {
			toggleEditComment($comment);
			$commentText.text($editCommentTextarea.val());
			$editCommentTextarea.val("");
		});
		return false;
	});
	$(".cancel-edit-comment").on("click", function(event) {
		var $comment= $(event.target).closest(".comment");
		toggleEditComment($comment);
		return false;
	});

	// New comment stuff.
	function toggleEditNewComment($newComment) {
		$newComment.find(".new-comment-body").toggle();
		$newComment.find(".new-comment-form").toggle();
	}
	var toggleNewComment = function(event) {
		var $newComment = $(event.target).closest(".new-comment");
		toggleEditNewComment($newComment);
		$newComment.find(".new-comment-text").focus();
		return false;
	};
	$(".new-comment-link").on("click", toggleNewComment);
	$(".cancel-new-comment").on("click", toggleNewComment);
	$(".new-comment-form").on("submit", function(event) {
		var $form = $(event.target);
		var $newComment = $form.closest(".new-comment");

		var data = {
			claimId: $newComment.closest(".claim").attr("claim-id"),
		};
		/*if ($form.find("input:checkbox[name='inContext']").is(":checked")) {
			data["contextClaimId"] = $(".bClaim").attr("claim-id");
		}*/
		submitForm($form, "/newComment/", data, function(r) {
			smartPageReload();
		});
		return false;
	});

	// New claim stuff.
	$(".new-claim-link").on("click", function(event) {
		$(this).tab("show");
		$(".new-claim-summary").focus();
		return false;
	});
	$(".new-claim-form").on("submit", function(event) {
		var $form = $(event.target);
		var data = {};
		submitForm($form, "/newClaim/", data, function(r) {
			smartPageReload();
		});
		return false;
	});

	// Add existing claim stuff.
	$(".add-existing-claim-link").on("click", function(event) {
		$(this).tab("show");
		$(".add-existing-claim-url").focus();
		return false;
	});
	$(".add-existing-claim-form").on("submit", function(event) {
		var $form = $(event.target);
		var data = {};
		submitForm($form, "/newInput/", data, function(r) {
			smartPageReload();
		});
		return false;
	});

	// Claim voting stuff.
	// voteClick is 1 is user clicked upvote and -1 if they clicked downvote.
	var processVote = function(voteClick, event) {
		var $target = $(event.target);
		var $vote = $target.closest(".vote");
		var $upvoteCount = $vote.find(".upvote-count");
		var $downvoteCount = $vote.find(".downvote-count");
		var currentVoteValue = +$vote.attr("my-vote");
		var newVoteValue = (voteClick === currentVoteValue) ? 0 : voteClick;
		var upvotes = +$upvoteCount.text();
		var downvotes = +$downvoteCount.text();

		// Update vote counts.
		// This logic has been created by looking at truth tables.
		if (currentVoteValue === 1) {
			upvotes -= 1;
		} else if (voteClick === 1) {
			upvotes += 1;
		}
		if (currentVoteValue === -1) {
			downvotes -= 1;
		} else if (voteClick === -1) {
			downvotes += 1;
		}
		$upvoteCount.text("" + upvotes);
		$downvoteCount.text("" + downvotes);

		// Update my-vote
		$vote.attr("my-vote", "" + newVoteValue);

		// Display my vote
		$vote.find(".upvote-link").toggleClass("on", newVoteValue === 1);
		$vote.find(".downvote-link").toggleClass("on", newVoteValue === -1);
		
		// Notify the server
		var data = {
			claimId: $target.closest(".claim").attr("claim-id"),
			value: newVoteValue,
		};
		$.ajax({
			type: 'POST',
			url: '/newVote/',
			data: JSON.stringify(data),
		})
		.done(function(r) {
		});
		return false;
	}
	$(".upvote-link").on("click", function(event) {
		return processVote(1, event);
	});
	$(".downvote-link").on("click", function(event) {
		return processVote(-1, event);
	});

	// Comment voting stuff.
	// voteClick is 1 is user clicked upvote and 0 if they clicked reset vote.
	$(".upvote-comment-link").on("click", function(event) {
		var $target = $(event.target);
		var $commentRow = $target.closest(".comment-row");
		var $upvoteCount = $commentRow.find(".comment-vote-count");

		// Update UI.
		$target.toggleClass("on");
		var newVoteValue = $target.hasClass("on") ? 1 : 0;
		var totalVotes = ((+$upvoteCount.text()) + (newVoteValue > 0 ? 1 : -1));
		if (totalVotes > 0) {
			$upvoteCount.text("" + totalVotes);
		} else {
			$upvoteCount.text("");
		}
		
		// Notify the server
		var data = {
			commentId: $commentRow.find(".comment").attr("comment-id"),
			value: newVoteValue,
		};
		$.ajax({
			type: 'POST',
			url: '/updateCommentVote/',
			data: JSON.stringify(data),
		})
		.done(function(r) {
		});
		return false;
	});

	// Subscription stuff.
	$(".subscribe-to-claim-link").on("click", function(event) {
		var $target = $(event.target);
		$target.toggleClass("on");
		var data = {
			claimId: $target.closest(".claim").attr("claim-id"),
		};
		$.ajax({
			type: 'POST',
			url: $target.hasClass("on") ? "/newSubscription/" : "/deleteSubscription/",
			data: JSON.stringify(data),
		})
		.done(function(r) {
		});
		return false;
	});
	$(".subscribe-comment-link").on("click", function(event) {
		var $target = $(event.target);
		$target.toggleClass("on");
		var data = {
			commentId: $target.closest(".comment-row").find(".comment").attr("comment-id"),
		};
		$.ajax({
			type: 'POST',
			url: $target.hasClass("on") ? "/newSubscription/" : "/deleteSubscription/",
			data: JSON.stringify(data),
		})
		.done(function(r) {
		});
		return false;
	});
});
