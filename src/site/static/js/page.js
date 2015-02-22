// This map contains page data we fetched from the server, e.g. when hovering over a intrasite link.
var fetchedPagesMap = {}; // pageId -> page data

// Send a new probability vote value to the server.
function newVote(value) {
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
	var html = converter.makeHtml(pageText);
	//html = html.replace(/\{\{p ([0-9]+)\}\}/g, "<a href='#' style='color:red'>$1</a>");
	var $pageText = $(".page-text")
	$pageText.html(html);
	$pageText.find("a").each(function(index, element) {
		var $element = $(element);
		var parts = $element.attr("href").match(/^(?:http:\/\/)?(?:www\.)?localhost:8012\/pages\/([0-9]+)\/?([0-9]+)?/)
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
		var $like = $target.closest(".like");
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
		$(event.target).closest(".my-vote").hide();
		var $voteSlider = $(".vote-slider");
		var $originalHandle = $voteSlider.find(".original");
		var $cloneHandle = $voteSlider.find(".clone");
		newVote(0.0);
		$(".my-vote-value").text("");
		$originalHandle.css("background-color", "#777777").css("left", $cloneHandle.css("left"));
		$cloneHandle.remove();
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
		$content.html($linkPopoverTemplate.html());
		$content.find(".like-count").text(page["LikeCount"]);
		$content.find(".dislike-count").text(page["DislikeCount"]);
		var myLikeValue = +page["MyLikeValue"];
		if (myLikeValue > 0) {
			$content.find(".disabled-like-link").addClass("on");
		} else if (myLikeValue < 0) {
			$content.find(".disabled-dislike-link").addClass("on");
		}
		if (page["Answers"] !== null) {
			$content.find(".vote").show();
			$content.find(".answer1").text(page["Answers"][0].Text);
			$content.find(".answer2").text(page["Answers"][1].Text);
			$content.find(".vote-text").text(page["VoteValue"] + "(" + page["VoteCount"] + ")");
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
					console.log("setting from cache");
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
						// TODO: handle deleted pages
						var page = JSON.parse(r);
						fetchedPagesMap[page["PageId"]] = page;
						var $popover = $("#" + $link.attr("aria-describedby"));
						var $content = $popover.find(".popover-content");
						console.log("setting from POST");
						$popover.find(".popover-title").text(page["Title"]);
						setPopoverContent($content, page);
						$link.popover("show");
					});
				}
				return '<img src="/static/images/loading.gif" class="loading-indicator" style="display:block"/>'
			}
	});
	$(".intrasite-link").on("shown.bs.popover", function () {
			
	})
});

// Initial setup.
$(function() {
	var pageType = $("body").attr("page-type");

	// Setup probability vote slider.
	if (pageType === "question") {
		var $myVote = $(".my-vote");
		var $myVoteValue = $(".my-vote-value");
		var $voteSlider = $(".vote-slider");
		var createVoteTick = function() {
			var $voteValue = $handle.clone();
			$voteValue.appendTo($voteSlider);
			$voteValue.css("background-color", "#777777").css("z-index", "0").addClass("clone");
			$handle.css("background-color", "").addClass("original");
		};
		$voteSlider.slider({
			min: 1,
			max: 99,
			step: 1,
			value: +$myVoteValue.text(),
			start: function(event, ui) {
				if (+$myVoteValue.text() <= 0) {
					$myVote.show();
					$myVoteValue.text(ui.value);
					createVoteTick();
				}
			},
			stop: function(event, ui) {
				newVote(+$myVoteValue.text());
			},
			slide: function(event, ui) {
				$myVoteValue.text(ui.value);
			},
		});
		var $handle = $voteSlider.find(".ui-slider-handle");
		if (+$myVoteValue.text() <= 0) {
			$handle.css("background-color", "#777777");
		} else {
			createVoteTick();
		}
	}
});
