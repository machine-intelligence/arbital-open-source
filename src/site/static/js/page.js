"use strict";

// This map contains page data we fetched from the server, e.g. when hovering over a intrasite link.
var fetchedPagesMap = {}; // pageId -> page data

// Send a new probability vote value to the server.
function postNewVote(value) {
	var data = {
		pageId: $("body").attr("page-id"),
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

// Format vote value into a pretty string.
// answers - array of two string elements corresponding to the answer text.
//   If undefined, the page's answers will be used.
function voteFormatter(value, answers) {
	if (value < 50) {
		var answer1Text = answers === undefined ? $("#answer1-text").text() : answers[0];
		return answer1Text + ": " + (100 - value) + "%";
	}
	var answer2Text = answers === undefined ? $("#answer2-text").text() : answers[1];
	return answer2Text + ": " + value + "%";
}

// Set the value of my vote.
function setMyVoteValue(valueStr) {
	myVoteValueStr = valueStr;
	$(".my-vote").toggle(valueStr !== "");
	$(".my-vote-value").text("| my vote is \"" + voteFormatter(+valueStr) + "\"");
}

// Setup vote slider behavior based on whether or not we voted.
function setupVoteSlider() {
	var mySlider = $("#vote-slider-input");
	if (myVoteValueStr !== "") {
		mySlider.bootstrapSlider("setValue", +myVoteValueStr);
	} else {
		var $voteSlider = $("#" + $("#vote-slider-input").attr("data-slider-id"));
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
			mySlider.bootstrapSlider("setValue", x);
		});
		$voteSlider.on("mouseup.z", function(event) {
			$voteSlider.off(".z");
		});
	}
}

// Replace Markdown text with corresponding HTML.
$(function() {
	// TODO: get pageText and add all the comment tags
	var converter = Markdown.getSanitizingConverter();
	/*converter.hooks.chain("preSpanGamut", function (text) {
		console.log("text: " + text);
		return text.replace(/(.*?)"""(.*?)"""(.*)/g, "$1<u>$2</u>$3");
	});*/
	/*console.log("LEN: " + pageText.length);
	converter.hooks.chain("preBlockGamut", function (text, rbg) {
		return text.replace(/.\n\n./g, function (whole, inner) {
			console.log("whole: " + whole);
			console.log("inner: " + inner);
			return whole[0] + "{{p " + inner + "}}\n\n" + whole[3];
		});
	});*/
	InitMathjax(converter, undefined, "");
	var html = converter.makeHtml(pageText);
	//html = html.replace(/\{\{p ([0-9]+)\}\}/g, "<a href='#' style='color:red'>$1</a>");
	var $pageText = $(".page-text")
	$pageText.html(html);

	// Setup attributes for links that are within our domain.
	var host = window.location.host;
	var re = new RegExp("^(?:https?:\/\/)?(?:www\.)?" + host + "\/pages\/([0-9]+)\/?([0-9]+)?");
	$pageText.find("a").each(function(index, element) {
		var $element = $(element);
		var parts = $element.attr("href").match(re);
		if (parts === null) return;
		$element.addClass("intrasite-link").attr("page-id", parts[1]).attr("privacy-key", parts[2]);
	});
});


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
	$(".delete-my-vote-link").on("click", function(event) {
		postNewVote(0.0);
		setMyVoteValue("");
		setupVoteSlider();
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
	var $linkPopoverTemplate = $("#link-popover-template");
	var setPopoverContent = function($content, page) {
		if (page["DeletedBy"] !== "0") {
			$content.html("");
			return;
		}
		$content.html($linkPopoverTemplate.html());
		$content.find(".like-count").text(page["LikeCount"]);
		$content.find(".dislike-count").text(page["DislikeCount"]);
		var myLikeValue = +page["MyLikeValue"];
		if (myLikeValue > 0) {
			$content.find(".disabled-like").addClass("on");
		} else if (myLikeValue < 0) {
			$content.find(".disabled-dislike").addClass("on");
		}
		if (page["Answers"] !== null) {
			$content.find(".vote").show();
			var answer1Text = page.Answers[0].Text;
			var answer2Text = page.Answers[1].Text;
			var answers = [answer1Text, answer2Text];
			if (page.VoteValue.Valid) {
				var x = Math.round(+page.VoteValue.Float64);
				answer1Text += " (" + (100 - x) + "%)";
				answer2Text += " (" + x + "%)";
			}
			$content.find(".answer1").text(answer1Text);
			$content.find(".answer2").text(answer2Text);
			$content.find(".vote-text").text(page.VoteValue + "(" + page.VoteCount + ")");
			var voteText = page.VoteCount + " vote" + (page.VoteCount === 1 ? "" : "s") + " counted";
			if (page["MyVoteValue"].Valid) {
				voteText += " | my vote is \"" + voteFormatter(+page.MyVoteValue.Float64, answers) + "\"";
			}
			$content.find(".vote-text").text(voteText);
		}
	}
	$(".intrasite-link").popover({ 
			html : true,
			placement: "top",
			trigger: "hover",
			delay: { "show": 500, "hide": 100 },
			title: function() {
				var pageId = $(this).attr("page-id");
				if (fetchedPagesMap[pageId]) {
					if (fetchedPagesMap[pageId]["DeletedBy"] !== "0") {
						return "[DELETED]";
					}
					return fetchedPagesMap[pageId]["Title"];
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
						fetchedPagesMap[page["PageId"]] = page;
						var $popover = $("#" + $link.attr("aria-describedby"));
						var $content = $popover.find(".popover-content");
						$popover.find(".popover-title").text(page["Title"]);
						setPopoverContent($content, page);
						$link.popover("show");
					});
				}
				return '<img src="/static/images/loading.gif" class="loading-indicator" style="display:block"/>'
			}
	});
});

// Initial setup.
$(function() {
	var pageType = $("body").attr("page-type");

	// Setup probability vote slider.
	if (pageType === "question") {
		var mySlider = $("#vote-slider-input").bootstrapSlider({
			step: 1,
			precision: 0,
			value: +myVoteValueStr,
			selection: "none",
			handle: "square",
			ticks: [1, 10, 20, 30, 40, 50, 60, 70, 80, 90, 99],
			formatter: voteFormatter,
		});
		var $voteSlider = $("#" + $("#vote-slider-input").attr("data-slider-id"));
		setupVoteSlider();

		// Update answer labels.
		if (voteCount > 0) {
			// Show the mean.
			var x = (voteValue - 1) * 100 / (99 - 1);
			var $voteTick = $voteSlider.find(".slider-tick").first().clone();
			$voteTick.addClass("vote-tick").css("left", x + "%");
			$voteSlider.find(".slider-track").append($voteTick);

			// Show the results next to answers.
			var vote = Math.round(voteValue);
			$("#answer1-vote-value").text("(" + (100 - vote) + "%)")
			$("#answer2-vote-value").text("(" + vote + "%)")
		}

		// Setup voting handlers.
		mySlider.bootstrapSlider("on", "slideStop", function(event){
			postNewVote(event.value);
			setMyVoteValue("" + event.value);
		});

		setMyVoteValue(myVoteValueStr);
	}
});
