function toggleEditInput($inputRight) {
	$inputRight.find(".inputBody").toggle();
	$inputRight.find(".editInputForm").toggle();
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

function toggleEditNewInput($bInput) {
	$bInput.find(".newInputLink").toggle();
	$bInput.find(".newInputForm").toggle();
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
			id: $(".bQuestion").attr("question-id"),
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
		$.each($form.serializeArray(), function(i, field) {
			data[field.name] = field.value;
		});

		toggleEditInput($inputRight);
		$inputRight.find(".inputUrl").attr("href", $inputRight.find(".editInputUrl").val());
		$inputRight.find(".inputText").text($inputRight.find(".editInputTextarea").val());

		$.ajax({
			type: 'POST',
			url: '/updateInput/',
			data: JSON.stringify(data),
		})
		.done(function(r) {
		});
		return false;
	});
	$(".cancelInput").on("click", function(event) {
		var $inputRight = $(event.target).closest(".inputRight");
		toggleEditInput($inputRight);
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
	var toggleNewComment = function(event) {
		var $newComment = $(event.target).closest(".newComment");
		toggleEditNewComment($newComment);
		$newComment.find(".inputNewComment").focus();
		return false;
	};
	$(".newCommentLink").on("click", toggleNewComment);
	$(".cancelNewComment").on("click", toggleNewComment);
	$(".saveNewComment").on("click", function(event) {
		var $newComment = $(event.target).closest(".newComment");
		var $inputNewComment = $newComment.find(".inputNewComment");
		var $newCommentText = $newComment.find(".newCommentText");
		var $parentComment = $newComment.closest(".comment");

		toggleEditNewComment($newComment);
		//$newCommentText.text($inputNewComment.val());
		//$inputNewComment.val("");

		var data = {
			inputId: $newComment.closest(".input").attr("input-id"),
			text: $inputNewComment.val(),
			questionId: $(".bQuestion").attr("question-id"),
		};
		if ($parentComment.length > 0) {
			data["replyToId"] = $parentComment.attr("comment-id");
		}
		$.ajax({
			type: 'POST',
			url: '/newComment/',
			data: JSON.stringify(data),
		})
		.done(function(r) {
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
		var $bInput = $form.closest(".bInput");
		var data = {};
		$.each($form.serializeArray(), function(i, field) {
			data[field.name] = field.value;
		});
		data["questionId"] = $(".bQuestion").attr("question-id");

		toggleEditNewInput($bInput);
		$.ajax({
			type: 'POST',
			url: '/newInput/',
			data: JSON.stringify(data),
		})
		.done(function(r) {
			$form[0].reset();
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
