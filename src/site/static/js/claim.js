function toggleEditClaim($claimRight) {
	$claimRight.find(".claimBody").toggle();
	$claimRight.find(".editClaimForm").toggle();
}

function toggleEditComment($commentBody) {
	$commentBody.toggle();
	$commentBody.siblings(".editCommentForm").toggle();
}

function toggleEditNewComment($newComment) {
	$newComment.find(".newCommentBody").toggle();
	$newComment.find(".newCommentForm").toggle();
}

function toggleEditNewClaim($newClaim) {
	$newClaim.find(".newClaimBody").toggle();
	$newClaim.find(".newClaimForm").toggle();
}

function toggleAddExistingClaim($newClaim) {
	$newClaim.find(".newClaimBody").toggle();
	$newClaim.find(".addExistingClaimForm").toggle();
}

// Reload the page with a lastVisit parameter so we can pretend that we are
// looking at a page at that time. This way new/updated markers are displayed
// correctly.
function smartPageReload() {
	var url = $("body").attr("claim-url");
	var lastVisit = encodeURIComponent($("body").attr("last-visit"));
	window.location.replace(url + "?lastVisit=" + lastVisit);
}

$(document).ready(function() {
	// Claim editing stuff.
	$(".editClaim").on("click", function(event) {
		var $claimRight = $(event.target).closest(".claimRight");
		var $claimTextarea = $claimRight.find(".editClaimTextarea");
		toggleEditClaim($claimRight);
		if ($claimTextarea.val() === "") {
			$claimTextarea.val($claimRight.find(".claimText").text());
			$claimRight.find(".editClaimUrl").val($claimRight.find(".claimUrl").attr("href"));
		}
		$claimTextarea.focus();
		return false;
	});
	$(".editClaimForm").on("submit", function(event) {
		var $form = $(event.target);
		var $claimRight = $(event.target).closest(".claimRight");

		var data = {};
		submitForm($form, "/updateClaim/", data, function(r) {
			var $claimUrl = $claimRight.find(".claimUrl");
			var url = $claimRight.find(".editClaimUrl").val();
			toggleEditClaim($claimRight);
			$claimUrl.attr("href", url);
			$claimUrl.toggle(url.length > 0);
			$claimRight.find(".claimText").text($claimRight.find(".editClaimTextarea").val());
		});
		return false;
	});
	$(".cancelEditClaim").on("click", function(event) {
		var $claimRight = $(event.target).closest(".claimRight");
		toggleEditClaim($claimRight);
		return false;
	});

	// Comment editing stuff.
	$(".editCommentLink").on("click", function(event) {
		var $commentBody = $(event.target).closest(".commentBody");
		var $form = $commentBody.siblings(".editCommentForm");
		var $editCommentTextarea = $form.find(".editCommentTextarea");
		var $commentText = $commentBody.find(".commentText");
		toggleEditComment($commentBody);
		if ($editCommentTextarea.val().length <= 0) {
			$editCommentTextarea.val($commentText.text());
		}
		$editCommentTextarea.focus();
		return false;
	});
	$(".editCommentForm").on("submit", function(event) {
		var $form = $(event.target);
		var $commentBody = $form.siblings(".commentBody");
		var $editCommentTextarea = $form.find(".editCommentTextarea");
		var $commentText = $commentBody.find(".commentText");

		var data = {id: $commentBody.closest(".comment").attr("comment-id")};
		submitForm($form, "/updateComment/", data, function(r) {
			toggleEditComment($commentBody);
			$commentText.text($editCommentTextarea.val());
			$editCommentTextarea.val("");
		});
		return false;
	});
	$(".cancelEditComment").on("click", function(event) {
		var $commentBody = $(event.target).closest(".editCommentForm").siblings(".commentBody");
		toggleEditComment($commentBody);
		return false;
	});

	// New comment stuff.
	var toggleNewComment = function(event) {
		var $newComment = $(event.target).closest(".newComment");
		toggleEditNewComment($newComment);
		$newComment.find(".newCommentTextarea").focus();
		return false;
	};
	$(".newCommentLink").on("click", toggleNewComment);
	$(".cancelNewComment").on("click", toggleNewComment);
	$(".newCommentForm").on("submit", function(event) {
		var $form = $(event.target);
		var $newComment = $form.closest(".newComment");
		var $parentComment = $newComment.closest(".comment");

		var data = {
			claimId: $newComment.closest(".claim").attr("claim-id"),
		};
		/*if ($form.find("input:checkbox[name='inContext']").is(":checked")) {
			data["contextClaimId"] = $(".bClaim").attr("claim-id");
		}*/
		if ($parentComment.length > 0) {
			data["replyToId"] = $parentComment.attr("comment-id");
		}
		submitForm($form, "/newComment/", data, function(r) {
			smartPageReload();
		});
		return false;
	});

	// New claim stuff.
	$(".newClaimLink").on("click", function(event) {
		var $newClaim = $(event.target).closest(".newClaim");
		toggleEditNewClaim($newClaim);
		$newClaim.find(".newClaimTextarea").focus();
		return false;
	});
	$(".newClaimForm").on("submit", function(event) {
		var $form = $(event.target);
		var data = {parentClaimId: $(".bClaim").attr("claim-id")};
		submitForm($form, "/newClaim/", data, function(r) {
			smartPageReload();
		});
		return false;
	});
	$(".cancelNewClaim").on("click", function(event) {
		var $newClaim = $(event.target).closest(".newClaim");
		toggleEditNewClaim($newClaim);
		return false;
	});

	// Add existing claim stuff.
	$(".addExistingClaimLink").on("click", function(event) {
		var $newClaim = $(event.target).closest(".newClaim");
		toggleAddExistingClaim($newClaim);
		$newClaim.find(".addExistingClaimForm").find("input:text[name='url']").focus();
		return false;
	});
	$(".addExistingClaimForm").on("submit", function(event) {
		var $form = $(event.target);
		var data = {parentClaimId: $(".bClaim").attr("claim-id")};
		submitForm($form, "/newInput/", data, function(r) {
			smartPageReload();
		});
		return false;
	});
	$(".cancelAddExistingClaim").on("click", function(event) {
		var $newClaim = $(event.target).closest(".newClaim");
		toggleAddExistingClaim($newClaim);
		return false;
	});

	// Claim voting stuff.
	// voteClick is 1 is user clicked upvote and -1 if they clicked downvote.
	var processVote = function(voteClick, event) {
		var $target = $(event.target);
		var $vote = $target.closest(".vote");
		var $upvoteCount = $vote.find(".upvoteCount");
		var $downvoteCount = $vote.find(".downvoteCount");
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
		$vote.find(".myUpvote").toggle(newVoteValue === 1);
		$vote.find(".myDownvote").toggle(newVoteValue === -1);
		
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
	$(".upvoteLink").on("click", function(event) {
		return processVote(1, event);
	});
	$(".downvoteLink").on("click", function(event) {
		return processVote(-1, event);
	});

	// Comment voting stuff.
	// voteClick is 1 is user clicked upvote and 0 if they clicked reset vote.
	var processCommentVote = function(newVoteValue, event) {
		var $commentBody = $(event.target).closest(".commentBody");
		var $upvoteCount = $commentBody.find(".commentVoteCount");
		var upvotes = +$upvoteCount.text();

		// Update UI.
		upvotes += newVoteValue > 0 ? 1 : -1;
		$upvoteCount.text("" + upvotes);
		$commentBody.find(".commentUpvoteLink").toggle();
		$commentBody.find(".commentResetVoteLink").toggle();
		
		// Notify the server
		var data = {
			commentId: $commentBody.closest(".comment").attr("comment-id"),
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
	}
	$(".commentUpvoteLink").on("click", function(event) {
		return processCommentVote(1, event);
	});
	$(".commentResetVoteLink").on("click", function(event) {
		return processCommentVote(0, event);
	});

	// Subscription stuff.
	$(".subscribeToClaim").on("click", function(event) {
		$(event.target).hide();
		$(".unsubscribeToClaim").show();
		var data = {
			claimId: $(".bClaim").attr("claim-id"),
		};
		$.ajax({
			type: 'POST',
			url: '/newSubscription/',
			data: JSON.stringify(data),
		})
		.done(function(r) {
		});
		return false;
	});
	$(".unsubscribeToClaim").on("click", function(event) {
		$(event.target).hide();
		$(".subscribeToClaim").show();
		var data = {
			claimId: $(".bClaim").attr("claim-id"),
		};
		$.ajax({
			type: 'POST',
			url: '/deleteSubscription/',
			data: JSON.stringify(data),
		})
		.done(function(r) {
		});
		return false;
	});
});
