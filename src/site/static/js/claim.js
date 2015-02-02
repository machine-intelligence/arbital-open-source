function toggleEditInput($inputRight) {
	$inputRight.find(".inputBody").toggle();
	$inputRight.find(".editInputForm").toggle();
}

function toggleEditQuestion($bQuestion) {
	$bQuestion.find(".editQuestionForm").toggle();
	$bQuestion.find(".questionBody").toggle();
}

function toggleEditComment($commentBody) {
	$commentBody.toggle();
	$commentBody.siblings(".editCommentForm").toggle();
}

function toggleEditNewComment($newComment) {
	$newComment.find(".newCommentBody").toggle();
	$newComment.find(".newCommentForm").toggle();
}

function toggleEditNewInput($bInput) {
	$bInput.find(".newInputLink").toggle();
	$bInput.find(".newInputForm").toggle();
}

$(document).ready(function() {
	// Question editing stuff.
	$(".editQuestion").on("click", function(event) {
		var $bQuestion = $(event.target).closest(".bQuestion");
		var $questionText = $bQuestion.find(".questionText");
		var $inputQuestion = $bQuestion.find(".inputQuestion");
		toggleEditQuestion($bQuestion);
		if ($inputQuestion.val().length <= 0) {
			$inputQuestion.val($questionText.text());
		}
		$inputQuestion.focus();
		return false;
	});
	$(".editQuestionForm").on("submit", function(event) {
		var $form = $(event.target);
		var $bQuestion = $form.closest(".bQuestion");
		var $questionText = $bQuestion.find(".questionText");
		var $inputQuestion = $bQuestion.find(".inputQuestion");

		var data = {id: $bQuestion.attr("question-id")};
		submitForm($form, "/updateQuestion/", data, function(r) {
			toggleEditQuestion($bQuestion);
			$questionText.text($inputQuestion.val());
			$inputQuestion.val("");
		});
		return false;
	});
	$(".cancelQuestion").on("click", function(event) {
		var $bQuestion = $(event.target).closest(".bQuestion");
		toggleEditQuestion($bQuestion);
		return false;
	});

	// Input editing stuff.
	$(".editInput").on("click", function(event) {
		var $inputRight = $(event.target).closest(".inputRight");
		var $inputTextarea = $inputRight.find(".editInputTextarea");
		toggleEditInput($inputRight);
		$inputRight.find(".editInputUrl").val($inputRight.find(".inputUrl").attr("href"));
		$inputTextarea.val($inputRight.find(".inputText").text());
		$inputTextarea.focus();
		return false;
	});
	$(".editInputForm").on("submit", function(event) {
		var $form = $(event.target);
		var $inputRight = $(event.target).closest(".inputRight");

		var data = {};
		submitForm($form, "/updateInput/", data, function(r) {
			var $inputUrl = $inputRight.find(".inputUrl");
			var url = $inputRight.find(".editInputUrl").val();
			toggleEditInput($inputRight);
			$inputUrl.attr("href", url);
			url.length <= 0 ? $inputUrl.hide() : $inputUrl.show();
			$inputRight.find(".inputText").text($inputRight.find(".editInputTextarea").val());
		});
		return false;
	});
	$(".cancelEditInput").on("click", function(event) {
		var $inputRight = $(event.target).closest(".inputRight");
		toggleEditInput($inputRight);
		return false;
	});

	// Comment editing stuff.
	$(".editCommentLink").on("click", function(event) {
		var $commentBody = $(event.target).closest(".commentBody");
		var $form = $commentBody.siblings(".editCommentForm");
		var $inputComment = $form.find(".inputComment");
		var $commentText = $commentBody.find(".commentText");
		toggleEditComment($commentBody);
		if ($inputComment.val().length <= 0) {
			$inputComment.val($commentText.text());
		}
		$inputComment.focus();
		return false;
	});
	$(".editCommentForm").on("submit", function(event) {
		var $form = $(event.target);
		var $commentBody = $form.siblings(".commentBody");
		var $inputComment = $form.find(".inputComment");
		var $commentText = $commentBody.find(".commentText");

		var data = {id: $commentBody.closest(".comment").attr("comment-id")};
		submitForm($form, "/updateComment/", data, function(r) {
			toggleEditComment($commentBody);
			$commentText.text($inputComment.val());
			$inputComment.val("");
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
		$newComment.find(".inputNewComment").focus();
		return false;
	};
	$(".newCommentLink").on("click", toggleNewComment);
	$(".cancelNewComment").on("click", toggleNewComment);
	$(".newCommentForm").on("submit", function(event) {
		var $form = $(event.target);
		var $newComment = $form.closest(".newComment");
		var $inputNewComment = $newComment.find(".inputNewComment");
		var $newCommentText = $newComment.find(".newCommentText");
		var $parentComment = $newComment.closest(".comment");

		var data = {
			inputId: $newComment.closest(".input").attr("input-id"),
			questionId: $(".bQuestion").attr("question-id"),
		};
		if ($parentComment.length > 0) {
			data["replyToId"] = $parentComment.attr("comment-id");
		}
		submitForm($form, "/newComment/", data, function(r) {
			location.reload();
		});
		return false;
	});

	// New input stuff.
	$(".newInputLink").on("click", function(event) {
		var $bInput = $(event.target).closest(".bInput");
		toggleEditNewInput($bInput);
		return false;
	});
	$(".newInputForm").on("submit", function(event) {
		var $form = $(event.target);
		var data = {questionId: $(".bQuestion").attr("question-id")};
		submitForm($form, "/newInput/", data, function(r) {
			location.reload();
		});
		return false;
	});
	$(".cancelNewInput").on("click", function(event) {
		var $bInput = $(event.target).closest(".bInput");
		toggleEditNewInput($bInput);
		return false;
	});

	// Voting stuff.
	$(".priorVote").on("click", function(event) {
		var $target = $(event.target);
		var data = {
			questionId: $(".bQuestion").attr("question-id"),
			value: "5.0",
		};
		$.ajax({
			type: 'POST',
			url: '/priorVote/',
			data: JSON.stringify(data),
		})
		.done(function(r) {
		});
		return false;
	});

	// Subscription stuff.
	$(".subscribeToQuestion").on("click", function(event) {
		$(event.target).hide();
		$(".unsubscribeToQuestion").show();
		var data = {
			questionId: $(".bQuestion").attr("question-id"),
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
	$(".unsubscribeToQuestion").on("click", function(event) {
		$(event.target).hide();
		$(".subscribeToQuestion").show();
		var data = {
			questionId: $(".bQuestion").attr("question-id"),
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
