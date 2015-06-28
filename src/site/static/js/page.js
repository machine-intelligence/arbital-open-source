"use strict";

// Create new PageJsController.
// page - page object corresponding to the page being displayed.
var PageJsController = function(page, $topParent, pageService, userService) {
	var page = page;
	var $topParent = $topParent;
	var pageId = page.PageId; // id of the page being displayed
	var userId = userService.user.Id;

	// This map contains page data we fetched from the server, e.g. when hovering over a intrasite link.
	// TODO: use pageService instead
	var fetchedPagesMap = {}; // pageId -> page data
	
	// Send a new probability vote value to the server.
	var postNewVote = function(pageId, value) {
		var data = {
			pageId: pageId,
			value: value,
		};
		$.ajax({
			type: "POST",
			url: "/newVote/",
			data: JSON.stringify(data),
		})
		.done(function(r) {
		});
	}
	
	// Set up a new vote slider. Set the slider's value based on the user's vote.
	var createVoteSlider = function($parent, pageId, votes, isPopoverVote) {
		// Convert votes into a user id -> {value, createdAt} map
		var voteMap = {};
		if (page.Votes) {
			for(var i = 0; i < page.Votes.length; i++) {
				var vote = page.Votes[i];
				voteMap[vote.UserId] = {value: vote.Value, createdAt: vote.CreatedAt};
			}
		}

		// Copy vote-template and add it to the parent.
		var $voteDiv = $("#vote-template").clone().show().attr("id", "vote" + pageId).appendTo($parent);
		var $input = $voteDiv.find(".vote-slider-input");
		$input.attr("data-slider-id", $input.attr("data-slider-id") + pageId);
		var userVoteStr = userId in voteMap ? ("" + voteMap[userId].value) : "";
		var mySlider = $input.bootstrapSlider({
			step: 1,
			precision: 0,
			selection: "none",
			handle: "square",
			value: +userVoteStr,
			ticks: [1, 10, 20, 30, 40, 50, 60, 70, 80, 90, 99],
			formatter: function(s) { return s + "%"; },
		});
		var $tooltip = $parent.find(".tooltip-main");

		// Set the value of the user's vote.
		var setMyVoteValue = function($voteDiv, userVoteStr) {
			$voteDiv.attr("my-vote", userVoteStr);
			$voteDiv.find(".my-vote").toggle(userVoteStr !== "");
			$voteDiv.find(".my-vote-value").text("| my vote is \"" + (+userVoteStr) + "%\"");
		}
		setMyVoteValue($voteDiv, userVoteStr);
	
		// Setup vote bars.
		// A bar represents users' votes for a given value. The tiled background
		// allows us to display each vote separately.
		var bars = {}; // voteValue -> {bar: jquery bar element, users: array of user ids who voted on this value}
		// Stuff for setting up the bars' css.
		var $barBackground = $parent.find(".bar-background");
		var $sliderTrack = $parent.find(".slider-track");
		var originLeft = $sliderTrack.offset().left;
		var originTop = $sliderTrack.offset().top;
		var barWidth = Math.max(5, $sliderTrack.width() / (99 - 1) * 2);
		// Set the correct css for the given bar object given the number of votes it has.
		var setBarCss = function(bar) {
			var $bar = bar.bar;
			var voteCount = bar.users.length;
			$bar.toggle(voteCount > 0);
			$bar.css("height", 11 * voteCount);
			$bar.css("z-index", 2 + voteCount);
			$barBackground.css("height", Math.max($barBackground.height(), $bar.height()));
			$barBackground.css("top", 10);
		}
		var highlightBar = function($bar, highlight) {
			var css = "url(/static/images/vote-bar.png)";
			var highlightColor = "rgba(128, 128, 255, 0.3)";
			if(highlight) {
				css = "linear-gradient(" + highlightColor + "," + highlightColor + ")," + css;
			}
			$bar.css("background", css);
			$bar.css("background-size", "100% 11px"); // have to set this each time
		};
		// Get the bar object corresponding to the given vote value. Create a new one if there isn't one.
		var getBar = function(vote) {
			if (!(vote in bars)) {
				var x = (vote - 1) / (99 - 1);
				var $bar = $("<div class='vote-bar'></div>");
				$bar.css("left", x * $sliderTrack.width() - barWidth / 2);
				$bar.css("width", barWidth);
				$barBackground.append($bar);
				bars[vote] = {bar: $bar, users: []};
			}
			return bars[vote];
		}
		for(var id in voteMap){
			// Create stacks for all the votes.
			var bar = getBar(voteMap[id].value);
			bar.users.push(id);
			setBarCss(bar);
		}

		// Convert mouse X position into % vote value.
		var voteValueFromMousePosX = function(mouseX) {
			var x = (mouseX - $sliderTrack.offset().left) / $sliderTrack.width();
			x = Math.max(0, Math.min(1, x));
			return Math.round(x * (99 - 1) + 1);
		};

		// Update the label that shows how many votes have been cast.
		var updateVoteCount = function() {
			var votesLength = Object.keys(voteMap).length;
			$voteDiv.find(".vote-count").text(votesLength + " vote" + (votesLength == 1 ? "" : "s"));
		};
		updateVoteCount();

		// Set handle's width.
		var $handle = $parent.find(".min-slider-handle");
		$handle.css("width", barWidth);

		// Don't track mouse movements and such for the vote in a popover.
		if (isPopoverVote) {
			if (!(userId in voteMap)) {
				$handle.hide();
			}
			return;
		}

		// Catch mousemove event on the text, so that it doesn't propagate to parent
		// and spawn popovers, which will prevent us clicking on "x" button to delete
		// our vote.
		$parent.find(".text-center").on("mousemove", function(event){
			return false;
		});

		var mouseInParent = false;
		var mouseInPopover = false;
		// Destroy the popover that displayes who voted for a given value.
		var $usersPopover = undefined;
		var destroyUsersPopover = function() {
			if ($usersPopover !== undefined) {
				$usersPopover.popover("destroy");
				highlightBar($usersPopover, false);
			}
			$usersPopover = undefined;
			mouseInPopover = false;
		};

		// Track mouse movement to show voter names.
		$parent.on("mouseenter", function(event) {
			mouseInParent = true;
			$handle.show();
			$tooltip.css("opacity", 1.0);
		});
		$parent.on("mouseleave", function(event) {
			mouseInParent = false;
			if (!(userId in voteMap)) {
				$handle.hide();
			} else {
				$input.bootstrapSlider("setValue", voteMap[userId].value);
			}
			$tooltip.css("opacity", 0.0);
			if (!mouseInPopover) {
				destroyUsersPopover();
			}
		});
		$parent.trigger("mouseleave");
		$parent.on("mousemove", function(event) {
			// Update slider.
			var voteValue = voteValueFromMousePosX(event.pageX);
			$input.bootstrapSlider("setValue", voteValue);
			if (mouseInPopover) return true;

			// We do a funky search to see if there is a vote nearby, and if so, show popover.
			var offset = 0, maxOffset = 5;
			var offsetSign = -1;
			while(offset <= maxOffset) {
				var hoverVoteValue = voteValue + offsetSign * offset;
				if (!(hoverVoteValue in bars)) {
					if(offsetSign < 0) offset++;
					offsetSign = -offsetSign;
					continue;
				}
				var bar = bars[hoverVoteValue];
				// Don't do anything if it's still the same bar as last time.
				if (bar.bar === $usersPopover) {
					break;
				}
				// Destroy old one.
				destroyUsersPopover();
				// Create new popover.
				$usersPopover = bar.bar;
				highlightBar(bar.bar, true);
				$usersPopover.popover({
					html : true,
					placement: "bottom",
					trigger: "manual",
					title: "Voters (" + hoverVoteValue + "%)",
					content: function() {
						var html = "";
						for(var i = 0; i < bar.users.length; i++) {
							var userId = bar.users[i];
							var user = userService.userMap[userId];
							var name = user.firstName + "&nbsp;" + user.lastName;
							html += "<a href='" + userService.getUserUrl(userId) + "'>" + name + "</a> " +
								"<span class='gray-text'>(" + voteMap[userId].createdAt + ")</span><br>";
						}
						return html;
					}
				}).popover("show");
				var $popover = $barBackground.find(".popover");
				$popover.on("mouseenter", function(event){
					mouseInPopover = true;
				});
				$popover.on("mouseleave", function(event){
					mouseInPopover = false;
					if (!mouseInParent) {
						destroyUsersPopover();
					}
				});
				break;
			}
			if (offset > maxOffset) {
				// We didn't find a bar, so destroy any existing popover.
				destroyUsersPopover();
			}
		});
	
		// Handle user's request to delete their vote.
		$voteDiv.find(".delete-my-vote-link").on("click", function(event) {
			var bar = bars[voteMap[userId].value];
			bar.users.splice(bar.users.indexOf(userId), 1);
			setBarCss(bar);
			if (bar.users.length <= 0){
				delete bars[voteMap[userId].value];
			}

			mouseInPopover = false;
			mouseInParent = false;
			delete voteMap[userId];
			$parent.trigger("mouseleave");
			$parent.trigger("mouseenter");

			postNewVote(pageId, 0.0);
			setMyVoteValue($voteDiv, "");
			updateVoteCount();
			return false;
		});
		
		// Track click to see when the user wants to vote / update their vote.
		$parent.on("click", function(event) {
			if (mouseInPopover) return true;
			if (userId in voteMap && voteMap[userId].value in bars) {
				// Update old bar.
				var bar = bars[voteMap[userId].value];
				bar.users.splice(bar.users.indexOf(userId), 1);
				setBarCss(bar);
				destroyUsersPopover();
			}

			// Set new vote and update all the things.
			var vote = voteValueFromMousePosX(event.pageX); 
			voteMap[userId] = {value: vote, createdAt: "now"};
			postNewVote(pageId, vote);
			setMyVoteValue($voteDiv, "" + vote);
			updateVoteCount();

			// Update new bar.
			var bar = getBar(vote);
			bar.users.push(userId);
			setBarCss(bar);
		});
	}
	
	// Add a popover to the given element. The element has to be an intrasite link jquery object.
	var setupIntrasiteLink = function($element) {
		var $linkPopoverTemplate = $("#link-popover-template");
		$element.popover({ 
			html : true,
			placement: "bottom",
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
				// TODO: replace this custom ajax fetching with our "standard" angularjs pageService.
				// Check if we already have this page cached.
				var page = fetchedPagesMap[pageId];
				if (page) {
					if (page.DeletedBy !== "0") {
						$content.html("");
						return "";
					}
					var $content = $("<div>" + $linkPopoverTemplate.html() + "</div>");
					$content.find(".popover-summary").text(page.Summary);
					$content.find(".like-count").text(page.LikeCount);
					$content.find(".dislike-count").text(page.DislikeCount);
					var myLikeValue = +page.MyLikeValue;
					if (myLikeValue > 0) {
						$content.find(".disabled-like").addClass("on");
					} else if (myLikeValue < 0) {
						$content.find(".disabled-dislike").addClass("on");
					}
					if (page.HasVote) {
						setTimeout(function(){
							var $popover = $("#" + $link.attr("aria-describedby"));
							var $content = $popover.find(".popover-content");
							createVoteSlider($content.find(".vote"), page.PageId, page.Votes, true);
						}, 100);
					}
					return $content.html();
				}
				// Check if we already issued a request to fetch this page.
				if (page === undefined) {
					// Fetch page data from the server.
					fetchedPagesMap[pageId] = null;
					var data = {pageAlias: pageId, privacyKey: $link.attr("privacy-key")};
					$.ajax({
						type: "POST",
						url: "/pageInfo/",
						data: JSON.stringify(data),
					})
					.success(function(r) {
						var page = JSON.parse(r);
						if (!page) return;
						fetchedPagesMap[page.PageId] = page;
						if (page.Alias && page.Alias !== page.PageId) {
							// Store the alias as well.
							fetchedPagesMap[page.Alias] = page;
						}
						$link.popover("show");
					});
				}
				return '<img src="/static/images/loading.gif" class="loading-indicator" style="display:block"/>'
			}
		});
	}

	// Highlight the page div. Used for selecting answers when #anchor matches.
	var highlightPageDiv = function() {
		$(".hash-anchor").removeClass("hash-anchor");
		$topParent.find(".page-body-div").addClass("hash-anchor");
	};
	if (window.location.hash === "#page-" + pageId) {
		highlightPageDiv();
	}
	
	// === Setup handlers.
	// Deleting a page
	$topParent.find(".delete-page-link").on("click", function(event) {
		$("#delete-page-alert").show();
		return false;
	});
	$topParent.find(".delete-page-cancel").on("click", function(event) {
		$("#delete-page-alert").hide();
	});
	$topParent.find(".delete-page-confirm").on("click", function(event) {
		var data = {
			pageId: pageId,
		};
		$.ajax({
			type: "POST",
			url: '/deletePage/',
			data: JSON.stringify(data),
		})
		.done(function(r) {
			smartPageReload();
		});
		return false;
	});
	
	// Page voting stuff.
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
			pageId: pageId,
			value: newLikeValue,
		};
		$.ajax({
			type: "POST",
			url: '/newLike/',
			data: JSON.stringify(data),
		})
		.done(function(r) {
		});
		return false;
	}
	$topParent.find(".like-link").on("click", function(event) {
		return processLike(1, event);
	});
	$topParent.find(".dislike-link").on("click", function(event) {
		return processLike(-1, event);
	});
	
	// Subscription stuff.
	$topParent.find(".subscribe-to-page-link").on("click", function(event) {
		var $target = $(event.target);
		$target.toggleClass("on");
		var data = {
			pageId: pageId,
		};
		$.ajax({
			type: "POST",
			url: $target.hasClass("on") ? "/newSubscription/" : "/deleteSubscription/",
			data: JSON.stringify(data),
		})
		.done(function(r) {
		});
		return false;
	});

	// Start initializes things that have to be killed when this editPage stops existing.
	this.start = function(pageVotes) {
		// Set up markdown.
		zndMarkdown.init(false, pageId, page.Text, $topParent);

		// Intrasite link hover.
		setupIntrasiteLink($topParent.find(".intrasite-link"));

		// Setup probability vote slider.
		if (page.HasVote) {
			createVoteSlider($topParent.find(".page-vote"), pageId, page.Votes, false);
		}
	};

	// Called before this controller is destroyed.
	this.stop = function() {
	};
};

// Directive for showing a standard Zanaduu page.
app.directive("zndPage", function (pageService, userService, $compile, $timeout) {
	return {
		templateUrl: "/static/html/page.html",
		controller: function ($scope, pageService, userService) {
			$scope.userService = userService;
			$scope.page = pageService.pageMap[$scope.pageId];
		},
		scope: {
			pageId: "=",
		},
		link: function(scope, element, attrs) {
			// Dynamically create comment elements.
			if (scope.page.Comments != null) {
				var $comments = element.find(".comments");
				for (var n = 0; n < scope.page.Comments.length; n++) {
					var $comment = $compile("<znd-comment page-id='\"" + scope.pageId +
						"\"' comment-index='" + n + "'></znd-comment>")(scope);
					$comments.prepend($comment);
				}
			}
			
			$timeout(function(){
				// Setup Page JS Controller.
				scope.pageJsController = new PageJsController(scope.page, element, pageService, userService);
				scope.pageJsController.start();
			});
		},
	};
});

// Directive for showing a comment.
app.directive("zndComment", function ($compile) {
	return {
		templateUrl: "/static/html/comment.html",
		controller: function ($scope, pageService, userService) {
			$scope.userService = userService;
			$scope.page = pageService.pageMap[$scope.pageId];
			$scope.comment = $scope.page.Comments[$scope.commentIndex];
			if ($scope.replyIndex !== undefined) {
				$scope.comment = $scope.comment.Replies[$scope.replyIndex];
			}

			var lastVisit = $scope.page.LastVisit;
			var showStars = $scope.comment.CreatorId != userService.user.Id && lastVisit != "";
			$scope.isNewComment = showStars && $scope.comment.CreatedAt >= lastVisit;
			$scope.isUpdatedComment = !$scope.isNewComment && showStars && $scope.comment.UpdatedAt >= lastVisit;
			if($scope.isNewComment) {
			console.log(lastVisit);
			console.log($scope.comment.CreatedAt);
			console.log($scope.comment.UpdatedAt);
			}
		},
		scope: {
			pageId: "=",
			commentIndex: "=",
			replyIndex: "=",
		},
		link: function(scope, element, attrs) {
			if (scope.replyIndex === undefined) {
				var $comments = element.find(".replies");
				// Dynamically create reply comment elements.
				if (scope.comment.Replies != null) {
					for (var n = 0; n < scope.comment.Replies.length; n++) {
						var $comment = $compile("<znd-comment page-id='\"" + scope.pageId +
							"\"' comment-index='" + scope.commentIndex +
							"' reply-index='" + n + "'></znd-comment>")(scope);
						$comments.append($comment);
					}
				}
				// Add New Comment element.
				var $newComment = $compile("<znd-new-comment page-id='\"" + scope.pageId +
						"\"' reply-to-id='\"" + scope.comment.Id + "\"'></znd-new-comment>")(scope);
				$comments.append($newComment);
			}

			// Comment voting stuff.
			// likeClick is 1 is user clicked like and 0 if they clicked reset like.
			element.find(".like-comment-link").on("click", function(event) {
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
					commentId: $commentRow.find(".comment-content").attr("comment-id"),
					value: newLikeValue,
				};
				$.ajax({
					type: "POST",
					url: '/updateCommentLike/',
					data: JSON.stringify(data),
				})
				.done(function(r) {
				});
				return false;
			});

			// Process comment subscribe click.
			element.find(".subscribe-comment-link").on("click", function(event) {
				var $target = $(event.target);
				$target.toggleClass("on");
				var data = {
					commentId: $target.closest(".comment-row").find(".comment-content").attr("comment-id"),
				};
				$.ajax({
					type: "POST",
					url: $target.hasClass("on") ? "/newSubscription/" : "/deleteSubscription/",
					data: JSON.stringify(data),
				})
				.done(function(r) {
				});
				return false;
			});
	
			// Comment editing stuff.
			function toggleEditComment($comment) {
				$comment.find(".comment-body").toggle();
				$comment.find(".edit-comment-form").toggle();
			}
			element.find(".edit-comment-link").on("click", function(event) {
				var $comment = $(event.target).closest(".comment-row").find(".comment-content");
				var $editCommentTextarea = $comment.find(".edit-comment-text");
				toggleEditComment($comment);
				$editCommentTextarea.focus();
				return false;
			});
			element.find(".edit-comment-form").on("submit", function(event) {
				var $form = $(event.target);
				var $comment = $form.closest(".comment-content");
				var $editCommentTextarea = $form.find(".edit-comment-text");
				var $commentText = $comment.find(".comment-text");
				var data = {id: $comment.attr("comment-id")};
				submitForm($form, "/updateComment/", data, function(r) {
					toggleEditComment($comment);
					$commentText.text($editCommentTextarea.val());
				});
				return false;
			});
			element.find(".cancel-edit-comment").on("click", function(event) {
				var $comment = $(event.target).closest(".comment-content");
				toggleEditComment($comment);
				return false;
			});
		},
	};
});

// Directive for creating a new comment.
app.directive("zndNewComment", function (pageService, userService) {
	return {
		templateUrl: "/static/html/newComment.html",
		controller: function ($scope, pageService, userService) {
		},
		scope: {
			pageId: "=",
			replyToId: "=",
		},
		link: function(scope, element, attrs) {
			// New comment stuff.
			var toggleNewComment = function(event) {
				var $newComment = $(event.target).closest(".new-comment");
				$newComment.find(".new-comment-body").toggle();
				$newComment.find(".new-comment-form").toggle();
				$newComment.find(".new-comment-text").focus();
				return false;
			};
			element.find(".new-comment-link").on("click", toggleNewComment);
			element.find(".cancel-new-comment").on("click", toggleNewComment);
			element.find(".new-comment-form").on("submit", function(event) {
				var $form = $(event.target);
				var data = {
					pageId: scope.pageId,
				};
				submitForm($form, "/newComment/", data, function(r) {
					smartPageReload();
				});
				return false;
			});
		},
	};
});
