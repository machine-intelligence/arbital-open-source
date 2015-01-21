function toggleEditSupport($support) {
	$support.find(".editSupport").toggle();
	$support.find(".supportText").toggle();
	$support.find(".saveSupport").toggle();
	$support.find(".cancelSupport").toggle();
	$support.find(".inputSupport").toggle();
}

function toggleEditQuestion() {
	$(".editQuestion").toggle();
	$(".questionText").toggle();
	$(".saveQuestion").toggle();
	$(".cancelQuestion").toggle();
	$(".inputQuestion").toggle();
}

function toggleEditComment($commentBody) {
	$commentBody.find(".editComment").toggle();
	$commentBody.find(".commentText").toggle();
}

function toggleEditNewComment($newComment) {
	$newComment.find(".newCommentLink").toggle();
	$newComment.find(".editNewComment").toggle();
}

$(document).ready(function() {
	// Question editing stuff.
	$(".editQuestion").on("click", function(event) {
		toggleEditQuestion();
		$(".inputQuestion").val($(".questionText").text());
		$(".inputQuestion").focus();
		return false;
	});
	$(".saveQuestion").on("click", function(event) {
		toggleEditQuestion();
		$(".questionText").text($(".inputQuestion").val());
		$(".inputQuestion").val("");

		var data = {
			id: $(".questionText").attr("question-id"),
			text: $(".questionText").text(),
		};
		$.ajax({
			type: 'POST',
			url: '/updateQuestion/',
			data: JSON.stringify(data),
		})
		.done(function(r) {
		});
		return false;
	});
	$(".cancelQuestion").on("click", function(event) {
		toggleEditQuestion();
		return false;
	});

	// Support editing stuff.
	$(".editSupport").on("click", function(event) {
		var $support = $(event.target).closest(".support");
		var $inputSupport = $support.find(".inputSupport");
		var $supportText = $support.find(".supportText");
		toggleEditSupport($support);
		$inputSupport.val($supportText.text());
		$inputSupport.focus();
		return false;
	});
	$(".saveSupport").on("click", function(event) {
		var $support = $(event.target).closest(".support");
		var $inputSupport = $support.find(".inputSupport");
		var $supportText = $support.find(".supportText");

		toggleEditSupport($support);
		$supportText.text($inputSupport.val());
		$inputSupport.val("");

		var data = {
			id: $support.attr("support-id"),
			text: $supportText.text(),
		};
		$.ajax({
			type: 'POST',
			url: '/updateSupport/',
			data: JSON.stringify(data),
		})
		.done(function(r) {
		});
		return false;
	});
	$(".cancelSupport").on("click", function(event) {
		var $support = $(event.target).closest(".support");
		toggleEditSupport($support);
		return false;
	});

	// Comment editing stuff.
	$(".commentText").on("click", function(event) {
		var $commentBody = $(event.target).closest(".commentBody");
		var $inputComment = $commentBody.find(".inputComment");
		var $commentText = $commentBody.find(".commentText");
		toggleEditComment($commentBody);
		$inputComment.val($commentText.text());
		$inputComment.focus();
		return false;
	});
	$(".saveComment").on("click", function(event) {
		var $commentBody = $(event.target).closest(".commentBody");
		var $inputComment = $commentBody.find(".inputComment");
		var $commentText = $commentBody.find(".commentText");

		toggleEditComment($commentBody);
		$commentText.text($inputComment.val());
		$inputComment.val("");

		var data = {
			id: $commentBody.closest(".comment").attr("comment-id"),
			text: $commentText.text(),
		};
		console.log(data);
		$.ajax({
			type: 'POST',
			url: '/updateComment/',
			data: JSON.stringify(data),
		})
		.done(function(r) {
		});
		return false;
	});
	$(".cancelComment").on("click", function(event) {
		var $commentBody = $(event.target).closest(".commentBody");
		toggleEditComment($commentBody);
		return false;
	});

	// New comment stuff.
	$(".newCommentLink").on("click", function(event) {
		var $newComment = $(event.target).closest(".newComment");
		toggleEditNewComment($newComment);
		return false;
	});
	$(".saveNewComment").on("click", function(event) {
		var $newComment = $(event.target).closest(".newComment");
		var $inputNewComment = $newComment.find(".inputNewComment");
		var $newCommentText = $newComment.find(".newCommentText");
		var $parentComment = $newComment.closest(".comment");

		toggleEditNewComment($newComment);
		//$newCommentText.text($inputNewComment.val());
		//$inputNewComment.val("");

		var data = {
			supportId: $newComment.closest(".support").attr("support-id"),
			text: $inputNewComment.val(),
		};
		if ($parentComment.length > 0) {
			data["replyToId"] = $parentComment.attr("comment-id");
		}
		$.ajax({
			type: 'POST',
			url: '/updateComment/',
			data: JSON.stringify(data),
		})
		.done(function(r) {
		});
		return false;
	});
	$(".cancelNewComment").on("click", function(event) {
		var $newComment = $(event.target).closest(".newComment");
		toggleEditNewComment($newComment);
		return false;
	});

	// Voting stuff.
	$(".priorVote").on("click", function(event) {
		var $target = $(event.target);
		var $support = $target.closest(".support");
		var data = {
			value: "5.0",
		};
		if ($target.attr("vote-id") === undefined) {
			data["supportId"] = $support.attr("support-id");
		} else {
			data["id"] = $target.attr("vote-id");
		}
		$.ajax({
			type: 'POST',
			url: '/updatePriorVote/',
			data: JSON.stringify(data),
		})
		.done(function(r) {
		});
		return false;
	});
	
});
