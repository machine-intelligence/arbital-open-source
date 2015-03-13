"use strict";

// This map contains page data we fetched from the server, e.g. when hovering over a intrasite link.
var fetchedPagesMap = {}; // pageId -> page data

// Send a new probability vote value to the server.
function postNewVote(pageId, value) {
	var data = {
		pageId: pageId,
		value: value,
	};
	$.ajax({
		type: 'POST',
		url: "/newVote/",
		data: JSON.stringify(data),
	})
	.done(function(r) {
	});
}

// Set the value of my vote.
function setMyVoteValue($voteDiv, valueStr) {
	$voteDiv.attr("my-vote", valueStr);
	$voteDiv.find(".my-vote").toggle(valueStr !== "");
	$voteDiv.find(".my-vote-value").text("| my vote is \"" + (+valueStr) + "%\"");
}

// Setup vote slider behavior based on whether or not we voted.
function setMyVoteOnSlider($input, myVoteValueStr) {
	var $voteSlider = $("#" + $input.attr("data-slider-id"));
	if (myVoteValueStr !== "") {
		$input.bootstrapSlider("setValue", +myVoteValueStr);
	} else {
		var $handle = $voteSlider.find(".min-slider-handle");
		$handle.hide();
		$voteSlider.on("mouseenter.z mouseleave.z", function(event) {
			$handle.toggle();
		});
		$voteSlider.on("mousemove.z", function(event) {
			var $track = $voteSlider.find(".slider-track");
			var x = (event.pageX - $track.offset().left) / $track.width();
			x = Math.max(0, Math.min(1, x));
			x = x * (99 - 1) + 1;
			$input.bootstrapSlider("setValue", x);
		});
		$voteSlider.on("mouseup.z", function(event) {
			$voteSlider.off(".z");
		});
	}
}

// Copy vote-template and add it to the parent. Set up a vote slider using the
// vote-slider-input insdie the cloned div. Set the slider to the given value.
function createVoteSlider($parent, pageId, voteCount, voteValueStr, myVoteValueStr) {
	var $voteDiv = $("#vote-template").clone().show().attr("id", "vote" + pageId).prependTo($parent);
	var $input = $voteDiv.find(".vote-slider-input");
	$input.attr("data-slider-id", $input.attr("data-slider-id") + pageId);
	var mySlider = $input.bootstrapSlider({
		step: 1,
		precision: 0,
		value: +myVoteValueStr,
		selection: "none",
		handle: "square",
		ticks: [1, 10, 20, 30, 40, 50, 60, 70, 80, 90, 99],
		formatter: function(s) { return s + "%"; },
	});
	var $voteSlider = $("#" + $input.attr("data-slider-id"));
	setMyVoteOnSlider($input, myVoteValueStr);

	// Show votes.
	if (voteCount > 0) {
		// Show the mean.
		var x = (+voteValueStr - 1) * 100 / (99 - 1);
		var $voteTick = $voteSlider.find(".slider-tick").first().clone();
		$voteTick.addClass("vote-tick").css("left", x + "%");
		$voteSlider.find(".slider-track").append($voteTick);
	}
	$voteDiv.find(".vote-count").text(voteCount + " vote" + (voteCount == 1 ? "" : "s"));

	// Setup voting handlers.
	mySlider.bootstrapSlider("on", "slideStop", function(event){
		postNewVote(pageId, event.value);
		setMyVoteValue($voteDiv, "" + event.value);
	});

	setMyVoteValue($voteDiv, myVoteValueStr);

	$voteDiv.find(".delete-my-vote-link").on("click", function(event) {
		postNewVote(pageId, 0.0);
		setMyVoteValue($voteDiv, "");
		setMyVoteOnSlider($input, "");
	});
}

// Set up markdown.
$(function() {
	setUpMarkdown(false);
});

// Add a popover to the given element. The element has to be an intrasite link jquery object.
function setupIntrasiteLink($element) {
	var $linkPopoverTemplate = $("#link-popover-template");
	var setPopoverContent = function($content, page) {
		if (page.DeletedBy !== "0") {
			$content.html("");
			return;
		}
		$content.html($linkPopoverTemplate.html());
		$content.find(".popover-summary").text(page.Summary);
		$content.find(".like-count").text(page.LikeCount);
		$content.find(".dislike-count").text(page.DislikeCount);
		var myLikeValue = +page.MyLikeValue;
		if (myLikeValue > 0) {
			$content.find(".disabled-like").addClass("on");
		} else if (myLikeValue < 0) {
			$content.find(".disabled-dislike").addClass("on");
		}
		if (page.Answers !== null) {
			$content.find(".vote").show();
			$content.find(".vote-text").text(page.VoteValue + "(" + page.VoteCount + ")");
			var voteText = page.VoteCount + " vote" + (page.VoteCount === 1 ? "" : "s") + " counted";
			if (page.MyVoteValue.Valid) {
				voteText += " | my vote is \"" + (+page.MyVoteValue.Float64) + "%\"";
			}
			$content.find(".vote-text").text(voteText);
		}
	}

	$element.popover({ 
		html : true,
		placement: "top",
		trigger: "hover",
		delay: { "show": 500, "hide": 100 },
		title: function() {
			var pageId = $(this).attr("page-id");
			if (fetchedPagesMap[pageId]) {
				if (fetchedPagesMap[pageId].DeletedBy !== "0") {
					return "[DELETED]";
				}
				return fetchedPagesMap[pageId].Title;
			}
			return "Loading...";
		},
		content: function() {
			var $link = $(this);
			var pageId = $link.attr("page-id");
			// Check if we already have this page cached.
			if (fetchedPagesMap[pageId]) {
				var $content = $($linkPopoverTemplate.html());
				setPopoverContent($content, fetchedPagesMap[pageId]);
				return $content.html();
			}
			// Check if we already issued a request to fetch this page.
			if (!(pageId in fetchedPagesMap)) {
				// Fetch page data from the server.
				fetchedPagesMap[pageId] = null;
				var data = {pageId: pageId, privacyKey: $link.attr("privacy-key")};
				$.ajax({
					type: "POST",
					url: "/pageInfo/",
					data: JSON.stringify(data),
				})
				.success(function(r) {
					var page = JSON.parse(r);
					fetchedPagesMap[page.PageId] = page;
					var $popover = $("#" + $link.attr("aria-describedby"));
					var $content = $popover.find(".popover-content");
					$popover.find(".popover-title").text(page.Title);
					setPopoverContent($content, page);
					$link.popover("show");
				});
			}
			return '<img src="/static/images/loading.gif" class="loading-indicator" style="display:block"/>'
		}
	});
}

// Setup handlers.
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
		var data = {};
		submitForm($form, "/updateClaim/", data, function(r) {
			smartPageReload();
		});
		return false;
	});

	// Deleting a page
	$(".delete-page-link").on("click", function(event) {
		$("#delete-page-alert").show();
		return false;
	});
	$(".delete-page-cancel").on("click", function(event) {
		$("#delete-page-alert").hide();
	});
	$(".delete-page-confirm").on("click", function(event) {
		var $target = $(event.target);
		var data = {
			pageId: $("body").attr("page-id"),
		};
		$.ajax({
			type: 'POST',
			url: '/deletePage/',
			data: JSON.stringify(data),
		})
		.done(function(r) {
			smartPageReload();
		});
		return false;
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
		var data = {
			pageId: $form.closest("body").attr("page-id"),
		};
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
	// likeClick is 1 is user clicked like and -1 if they clicked dislike.
	var processLike = function(likeClick, event) {
		var $target = $(event.target);
		var $like = $target.closest(".page-like-div");
		var $likeCount = $like.find(".like-count");
		var $dislikeCount = $like.find(".dislike-count");
		var currentLikeValue = +$like.attr("my-like");
		var newLikeValue = (likeClick === currentLikeValue) ? 0 : likeClick;
		var likes = +$likeCount.text();
		var dislikes = +$dislikeCount.text();

		// Update like counts.
		// This logic has been created by looking at truth tables.
		if (currentLikeValue === 1) {
			likes -= 1;
		} else if (likeClick === 1) {
			likes += 1;
		}
		if (currentLikeValue === -1) {
			dislikes -= 1;
		} else if (likeClick === -1) {
			dislikes += 1;
		}
		$likeCount.text("" + likes);
		$dislikeCount.text("" + dislikes);

		// Update my-like
		$like.attr("my-like", "" + newLikeValue);

		// Display my like
		$like.find(".like-link").toggleClass("on", newLikeValue === 1);
		$like.find(".dislike-link").toggleClass("on", newLikeValue === -1);
		
		// Notify the server
		var data = {
			pageId: $target.closest("body").attr("page-id"),
			value: newLikeValue,
		};
		$.ajax({
			type: 'POST',
			url: '/newLike/',
			data: JSON.stringify(data),
		})
		.done(function(r) {
		});
		return false;
	}
	$(".like-link").on("click", function(event) {
		return processLike(1, event);
	});
	$(".dislike-link").on("click", function(event) {
		return processLike(-1, event);
	});

	// Comment voting stuff.
	// likeClick is 1 is user clicked like and 0 if they clicked reset like.
	$(".like-comment-link").on("click", function(event) {
		var $target = $(event.target);
		var $commentRow = $target.closest(".comment-row");
		var $likeCount = $commentRow.find(".comment-like-count");

		// Update UI.
		$target.toggleClass("on");
		var newLikeValue = $target.hasClass("on") ? 1 : 0;
		var totalLikes = ((+$likeCount.text()) + (newLikeValue > 0 ? 1 : -1));
		if (totalLikes > 0) {
			$likeCount.text("" + totalLikes);
		} else {
			$likeCount.text("");
		}
		
		// Notify the server
		var data = {
			commentId: $commentRow.find(".comment").attr("comment-id"),
			value: newLikeValue,
		};
		$.ajax({
			type: 'POST',
			url: '/updateCommentLike/',
			data: JSON.stringify(data),
		})
		.done(function(r) {
		});
		return false;
	});

	// Subscription stuff.
	$(".subscribe-to-page-link").on("click", function(event) {
		var $target = $(event.target);
		$target.toggleClass("on");
		var data = {
			pageId: $target.closest("body").attr("page-id"),
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

	// Intrasite link hover.
	setupIntrasiteLink($(".intrasite-link"));
});

// Initial setup.
$(function() {
	// Setup probability vote slider.
	if ($("body").attr("has-vote") !== "") {
		createVoteSlider($("#main-page-vote"), $("body").attr("page-id"), gVoteCount, gVoteValueStr, gMyVoteValueStr);
	}
});
