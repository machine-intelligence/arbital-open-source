'use strict';

/* jasmine specs for directives go here */

describe('directives', function() {

	var $compile,
			$rootScope;
	var scope, ctrl, $httpBackend;

	// Load the myApp module, which contains the directive
	beforeEach(module('arbital'));
	beforeEach(module('templates'));

	var testPage = {

		pageId:1,
		edit:0,
		type:"",
		title:"title",
		clickbait:"",
		textLength:0,
		alias:"",
		sortChildrenBy:"",
		hasVote:false,
		voteType:"",
		creatorId:0,
		createdAt:"",
		originalCreatedAt:"",
		editKarmaLock:0,
		seeGroupId:0,
		editGroupId:0,
		isAutosave:false,
		isSnapshot:false,
		isCurrentEdit:false,
		isMinorEdit:false,
		todoCount:0,
		anchorContext:"",
		anchorText:"",
		anchorOffset:0,

		text:"text",
		metaText:"",

		isSubscribed:false,
		subscriberCount:0,
		likeCount:0,
		dislikeCount:0,
		myLikeValue:0,
		likeScore:0,
		lastVisit:"",
		hasDraft:false,

		currentEditNum:0,
		wasPublished:false,
		votes:[],
		lockedVoteType:"",
		maxEditEver:0,
		myLastAutosaveEdit:0,
		redLinkCount:0,
		childDraftId:0,
		lockedBy:0,
		lockedUntil:"",
		nextPageId:0,
		prevPageId:0,
		usedAsMastery:false,

		summaries:[],

		answerIds:[],
		commentIds:[],
		questionIds:[],
		lensIds:[],
		taggedAsIds:[],
		relatedIds:[],
		requirementIds:[],
		subjectIds:[],

		answerCount:0,
		commentCount:0,

		domainIds:[],

		changeLogs:[],

		hasChildren:false,
		hasParents:false,
		childIds:[],
		parentIds:[],

		members:[]
	};

	var testUser = {
		id:"1",
		firstName:"firstName",
		lastName:"lastname",
		lastWebsiteVisit:0,
		isSubscribed:0,
	}

	// Store references to $rootScope and $compile
	// so they are available to all tests in this describe block
	beforeEach(inject(function(_$compile_, _$rootScope_){
		// The injector unwraps the underscores (_) from around the parameter names when matching
		$compile = _$compile_;
		$rootScope = _$rootScope_;
	}));

	beforeEach(inject(function(_$httpBackend_, $rootScope, $controller) {
		$httpBackend = _$httpBackend_;

		$httpBackend.whenPOST('/json/userPopover/').
				respond([{}]);

		$httpBackend.whenPOST('/json/intrasitePopover/').
				respond([{}]);

		$httpBackend.whenGET('static/icons/arbital-logo.svg').respond([{}]);
		$httpBackend.whenGET('static/icons/thumb-up-outline.svg').respond([{}]);
		$httpBackend.whenGET('static/icons/thumb-down-outline.svg').respond([{}]);
		$httpBackend.whenGET('static/icons/link-variant.svg').respond([{}]);
		$httpBackend.whenGET('static/icons/comment-plus-outline.svg').respond([{}]);
		$httpBackend.whenGET('static/icons/format-header-pound.svg').respond([{}]);

		ctrl = $controller('ArbitalCtrl', {$scope: $rootScope});

		$rootScope.pageService.addPageToMap(testPage);
		$rootScope.pageId = 1;
		$rootScope.pageService.primaryPage = testPage;
		$rootScope.pageService.editMap[$rootScope.pageId] = testPage;
		$rootScope.userService.user = testUser;
	}));
/*
	it('testing arb-user-popover', function() {
		var element = $compile("<arb-user-popover user-id='" + "1" +
			"' direction='" + "down" + "' arrow-offset='" + "0" +
			"'></arb-user-popover>")($rootScope);
		$rootScope.$digest();
		console.log(element);
		expect(element.html()).toNotEqual("");
	});

	it('testing arb-intrasite-popover', function() {
		var element = $compile("<arb-intrasite-popover page-id='" + 1 +
			"' direction='" + "down" + "' arrow-offset='" + 0 +
			"'></arb-intrasite-popover>")($rootScope);
		$rootScope.$digest();
		console.log(element);
		expect(element.html()).toNotEqual("");
	});

	it('testing arb-edit-page', function() {
		var element = $compile("<arb-edit-page class='full-height' page-id='" + 1 +
			"' done-fn='doneFn(result)' layout='column'></arb-edit-page>")($rootScope);
		$rootScope.$digest();
		console.log(element);
		expect(element.html()).toNotEqual("");
	});

	it('testing arb-inline-comment', function() {
		var element = $compile($("<arb-inline-comment" +
			" lens-id='" + 1 +
			"' comment-id='" + 1 + "'></arb-inline-comment>"))($rootScope);
		$rootScope.$digest();
		console.log(element);
		expect(element.html()).toNotEqual("");
	});

	it('testing arb-discussion', function() {
		var element = $compile("<arb-discussion class='reveal-after-render' page-id='" + 1 +
			"'></arb-discussion>")($rootScope);
		$rootScope.$digest();
		console.log(element);
		expect(element.html()).toNotEqual("");
	});

/*
	it('testing arb-group-index', function() {
		var element = $compile("<arb-group-index group-id='" + 1 +
			"' ids-map='indexPageIdsMap'></arb-group-index>")($rootScope);
		$rootScope.$digest();
		console.log(element);
		expect(element.html()).toNotEqual("");
	});
*/
/*
	it('testing arb-index', function() {
		testPage.pageId = 3440973961008233681;

		var element = $compile("<arb-index featured-domains='featuredDomains'></arb-index>")($rootScope);

		$rootScope.$digest();
		console.log(element);
		expect(element.html()).toContain("a");
	});
*/
/*
	it('testing arb-group-index', function() {
		var element = $compile("<arb-primary-page></arb-primary-page>")($rootScope);

		$rootScope.$digest();
		console.log(element);
		expect(element.html()).toContain("a");
	});
*/

	function compileElement(elementText) {
		console.log(testPage.text);
		var element = $compile(elementText)($rootScope);
		$rootScope.$digest();
		console.log(element.html());
		return element;
	}

	// Perform a test, expecting an address tag as the result.
	// Returns the address tag jQuery variable for further processing
	// options {
	//	 expectTextToEqual: the expected exact contents of the address tag's text
	//	 expectTextToContain[]: an array of text that the address tag's text is expected to contain
	//	 expectTextToNotContain[]: an array of text that the address tag's text is expected to not contain
	//	 expectHrefToEqual: the expected exact contents of the address tag's href
	//	 expectHrefToContain[]: an array of text that the address tag's href is expected to contain
	//	 expectHrefToNotContain[]: an array of text that the address tag's href is expected to not contain
	//	 expectClassToContain[]: an array of text that the address tag's class is expected to contain
	//	 expectClassToNotContain[]: an array of text that the address tag's class is expected to not contain
	//	 expectPageIdToEqual: the expected contents of the page-id attribute
	//	 expectUserIdToEqual: the expected contents of the user-id attribute
	//	 expectEmbedVoteIdToEqual: the expected contents of the embed-vote-id attribute
	// }
	function expectAddressTag(pageText, options) {
		var options = options || {};

		testPage.text = pageText;
		var element = compileElement(elementText);
		var $aTag = $(element.html()).find("a");

		if (options.expectTextToEqual) {
			expect($aTag.text()).toEqual(options.expectTextToEqual);
		}
		if (options.expectTextToContain) {
			for (var index in options.expectTextToContain) {
				expect($aTag.text()).toContain(options.expectTextToContain[index]);
			}
		}
		if (options.expectTextToNotContain) {
			for (var index in options.expectTextToNotContain) {
				expect($aTag.text()).toNotContain(options.expectTextToNotContain[index]);
			}
		}
		if (options.expectHrefToEqual) {
			expect($aTag.attr("href")).toEqual(options.expectHrefToEqual);
		}
		if (options.expectHrefToContain) {
			for (var index in options.expectHrefToContain) {
				expect($aTag.attr("href")).toContain(options.expectHrefToContain[index]);
			}
		}
		if (options.expectHrefToNotContain) {
			for (var index in options.expectHrefToNotContain) {
				expect($aTag.attr("href")).toNotContain(options.expectHrefToNotContain[index]);
			}
		}
		if (options.expectClassToContain) {
			for (var index in options.expectClassToContain) {
				expect($aTag.attr("class")).toContain(options.expectClassToContain[index]);
			}
		}
		if (options.expectClassToNotContain) {
			for (var index in options.expectClassToNotContain) {
				expect($aTag.attr("class")).toNotContain(options.expectClassToNotContain[index]);
			}
		}
		if (options.expectPageIdToEqual) {
			expect($aTag.attr("page-id")).toEqual(options.expectPageIdToEqual);
		}
		if (options.expectUserIdToEqual) {
			expect($aTag.attr("user-id")).toEqual(options.expectUserIdToEqual);
		}
		if (options.expectEmbedVoteIdToEqual) {
			expect($aTag.attr("embed-vote-id")).toEqual(options.expectEmbedVoteIdToEqual);
		}
		return $aTag;
	}

	// Perform a test, expecting a paragraph tag as the result.
	// Returns the paragraph tag jQuery variable for further processing
	// options {
	//	 expectTextToEqual: the expected contents of the paragraph tag's text
	//	 expectTextToContain[]: an array of text that the paragraph tag's text is expected to contain
	//	 expectTextToNotContain[]: an array of text that the paragraph tag's text is expected to not contain
	// }
	function expectParagraphTag(pageText, options) {
		var options = options || {};

		testPage.text = pageText;
		var element = compileElement(elementText);
		var $pTag = $(element.html());

		if (options.expectTextToEqual) {
			expect($pTag.text()).toEqual(options.expectTextToEqual);
		}
		if (options.expectTextToContain) {
			for (var index in options.expectTextToContain) {
				expect($pTag.text()).toContain(options.expectTextToContain[index]);
			}
		}
		if (options.expectTextToNotContain) {
			for (var index in options.expectTextToNotContain) {
				expect($pTag.text()).toNotContain(options.expectTextToNotContain[index]);
			}
		}
		return $pTag;
	}

	// Perform a test, expecting the result to be completely empty
	function expectEmptyElement(pageText) {
		testPage.text = pageText;
		var element = compileElement(elementText);
		expect(element.text()).toEqual("");
	}

	var elementText = "<arb-markdown class='popover-text-container' page-id='1'></arb-markdown>";

	it('testing markdown', function() {
		var testPage2 = {
			pageId:2,
			alias:"existentPageAlias",
			title:"existentPageTitle",
			seeGroupId:"0"
		};
		$rootScope.pageService.addPageToMap(testPage2);

		expectAddressTag("[existentPageAlias]", {expectTextToEqual:"ExistentPageTitle", expectClassToNotContain:["red-link"], expectPageIdToEqual:"2"});
		expectAddressTag("[nonexistentPageAlias]", {expectTextToEqual:"nonexistentPageAlias", expectClassToContain:["red-link"], expectPageIdToEqual:"nonexistentPageAlias"});
		expectAddressTag("[existentPageAlias description]", {expectTextToEqual:"description", expectClassToNotContain:["red-link"], expectPageIdToEqual:"2"});
		expectAddressTag("[nonexistentPageAlias description]", {expectTextToEqual:"description", expectClassToContain:["red-link"], expectPageIdToEqual:"nonexistentPageAlias"});
		expectParagraphTag("[hyphenated-alias]", {expectTextToEqual:"[hyphenated-alias]"});
		expectParagraphTag("[hyphenated-alias description]", {expectTextToEqual:"[hyphenated-alias description]"});
		expectParagraphTag("[^%@#&^!@ test]", {expectTextToEqual:"[^%@#&^!@ test]"});
		expectAddressTag("[http://google.com google]", {expectTextToEqual:"google", expectHrefToEqual:"http://google.com"});
		expectAddressTag("[ text]", {expectTextToEqual:"text", expectHrefToContain:"/edit", expectClassToContain:["red-link"], expectPageIdToEqual:"0"});
		expectAddressTag("[@1]", {expectTextToEqual:"title", expectHrefToContain:"/user/1", expectClassToNotContain:["red-link"], expectUserIdToEqual:"1"});
		expectAddressTag("[@999]", {expectTextToEqual:"999",expectHrefToContain:["/user/999"],expectClassToContain:["red-link"],expectUserIdToEqual:"999"});
		expectParagraphTag("[text](existentPageAlias)", {expectTextToEqual:"text"});
		expectParagraphTag("[text](nonexistentPageAlias)", {expectTextToEqual:"text"});
		expectAddressTag("[text](http://google.com)", {expectTextToEqual:"text",expectHrefToEqual:"http://google.com"});
		expectAddressTag("[vote:existentPageAlias]", {expectTextToContain:["Embedded existentPageAlias vote."],expectHrefToContain:["/pages/existentPageAlias/?embedVote=1"],expectPageIdToEqual:"existentPageAlias",expectEmbedVoteIdToEqual:"existentPageAlias"});
		expectAddressTag("[vote:nonexistentPageAlias]", {expectTextToContain:["Embedded nonexistentPageAlias vote."],expectHrefToContain:["/pages/nonexistentPageAlias/?embedVote=1"],expectPageIdToEqual:"nonexistentPageAlias",expectEmbedVoteIdToEqual:"nonexistentPageAlias"});
		expectParagraphTag("[todo:text]", {expectTextToEqual:""});
		expectParagraphTag("[comment:text]", {expectTextToEqual:""});
		expectEmptyElement("[summary(optional):markdown]");
		expectAddressTag("[text](http://foo.com/blah_(wikipedia)#cite-1)", {expectTextToEqual:"text",expectHrefToEqual:"http://foo.com/blah_(wikipedia)#cite-1"});
		expectAddressTag("[text](http://www.example.com/wpstyle/?p=364)", {expectTextToEqual:"text",expectHrefToEqual:"http://www.example.com/wpstyle/?p=364"});
		expectAddressTag("[text](https://www.example.com/foo/?bar=baz&inga=42&quux)", {expectTextToEqual:"text",expectHrefToEqual:"https://www.example.com/foo/?bar=baz&inga=42&quux"});
		expectAddressTag("[text](http://userid:password@example.com:8080)", {expectTextToEqual:"text",expectHrefToEqual:"http://userid:password@example.com:8080"});
		expectAddressTag("[text](http://foo.bar/?q=Test%20URL-encoded%20stuff)", {expectTextToEqual:"text",expectHrefToEqual:"http://foo.bar/?q=Test%20URL-encoded%20stuff"});
		expectParagraphTag("[text](http://)", {expectTextToEqual:"text"});
		expectParagraphTag("[text]()", {expectTextToEqual:"text"});
		expectParagraphTag("\\[existentPageAlias]", {expectTextToEqual:"[existentPageAlias]"});
		expectParagraphTag("[existentPageAlias\\]", {expectTextToEqual:"[existentPageAlias]"});
		expectParagraphTag("\\[existentPageAlias\\]", {expectTextToEqual:"[existentPageAlias]"});
		expectAddressTag("\\\\[existentPageAlias]", {expectTextToEqual:"ExistentPageTitle",expectClassToNotContain:["red-link"],expectPageIdToEqual:"2"});
		expectParagraphTag("[existentPageAlias\\\\]", {expectTextToEqual:"[existentPageAlias\\]"});
		expectParagraphTag("\\\\[existentPageAlias\\\\]", {expectTextToEqual:"\\[existentPageAlias\\]"});
		expectParagraphTag("\\[vote:existentPageAlias]", {expectTextToEqual:"[vote:existentPageAlias]"});
		expectParagraphTag("[vote:existentPageAlias\\]", {expectTextToEqual:"[vote:existentPageAlias]"});
		expectParagraphTag("\\[vote:existentPageAlias\\]", {expectTextToEqual:"[vote:existentPageAlias]"});
		expectAddressTag("\\\\[vote:existentPageAlias]", {expectTextToContain:["Embedded existentPageAlias vote."],expectHrefToContain:["/pages/existentPageAlias/?embedVote=1"],expectPageIdToEqual:"existentPageAlias",expectEmbedVoteIdToEqual:"existentPageAlias"});
		expectParagraphTag("[vote:existentPageAlias\\\\]", {}); //expectTextToEqual:"[vote:existentPageAlias\\]"
		expectParagraphTag("\\\\[vote:existentPageAlias\\\\]", {}); //expectTextToEqual:"\\[vote:existentPageAlias\\]"
		expectAddressTag("\\[text](http://google.com)", {}); //expectTextToEqual:"http://google.com", expectHrefToEqual:"http://google.com"
		expectAddressTag("[text\\](http://google.com)", {}); //expectTextToEqual:"http://google.com", expectHrefToEqual:"http://google.com"
		expectAddressTag("[text]\\(http://google.com)", {}); //expectTextToEqual:"text", expectHrefToContain:["/edit/text"]
		expectAddressTag("[text](http://google.com\\)", {}); //expectTextToEqual:"http://google.com)", expectHrefToEqual:"http://google.com)"
		expectAddressTag("\\\\[text](http://google.com)", {expectTextToEqual:"text",expectHrefToEqual:"http://google.com"});
		expectAddressTag("[text\\\\](http://google.com)", {}); //expectTextToEqual:"text\\", expectHrefToEqual:"http://google.com"
		expectAddressTag("[text]\\\\(http://google.com)", {}); //expectTextToEqual:"texthttp://google.com", expectHrefToContain:["/edit/text"]
		expectParagraphTag("[text](http://google.com\\\\)", {}); //expectTextToEqual:"text"
		expectParagraphTag("\\[@1]", {expectTextToEqual:"[@1]"});
		expectParagraphTag("[@1\\]", {expectTextToEqual:"[@1]"});
		expectParagraphTag("\\[@1\\]", {expectTextToEqual:"[@1]"});
		expectAddressTag("\\\\[@1]", {expectTextToEqual:"title",expectHrefToContain:["/user/1"],expectClassToNotContain:["red-link"],expectUserIdToEqual:"1"});
		expectParagraphTag("[@1\\\\]", {expectTextToEqual:"[@1\\]"});
		expectParagraphTag("\\\\[@1\\\\]", {expectTextToEqual:"\\[@1\\]"});
		expectParagraphTag("\\[ text]", {expectTextToEqual:"[ text]"});
		expectAddressTag("[ text\\]", {}); //expectTextToEqual:"http://arbital.com/edit", expectHrefToContain:["/edit"]
		expectParagraphTag("\\[ text\\]", {expectTextToEqual:"[ text]"});
		expectAddressTag("\\\\[ text]", {expectTextToEqual:"text",expectHrefToContain:["/edit"]});
		expectAddressTag("[ text\\\\]", {}); //expectTextToEqual:"text\\", expectHrefToContain:["/edit"]
		expectAddressTag("\\\\[ text\\\\]", {}); //expectTextToEqual:"text\\", expectHrefToContain:["/edit"]
		expectAddressTag("[ExistentPageAlias]", {expectTextToEqual:"ExistentPageTitle",expectClassToNotContain:["red-link"],expectPageIdToEqual:"2"});
		expectAddressTag("[NonexistentPageAlias]", {expectTextToEqual:"NonexistentPageAlias",expectClassToContain:["red-link"],expectPageIdToEqual:"NonexistentPageAlias"});
		expectAddressTag("[-ExistentPageAlias]", {expectTextToEqual:"existentPageTitle",expectClassToNotContain:["red-link"],expectPageIdToEqual:"2"});
		expectAddressTag("[-NonexistentPageAlias]", {expectTextToEqual:"NonexistentPageAlias",expectClassToContain:["red-link"],expectPageIdToEqual:"NonexistentPageAlias"});
		expectAddressTag("[existentPageAlias]", {expectTextToEqual:"ExistentPageTitle",expectClassToNotContain:["red-link"],expectPageIdToEqual:"2"});
		expectAddressTag("[nonexistentPageAlias]", {expectTextToEqual:"nonexistentPageAlias",expectClassToContain:["red-link"],expectPageIdToEqual:"nonexistentPageAlias"});
		expectAddressTag("[-existentPageAlias]", {expectTextToEqual:"existentPageTitle",expectClassToNotContain:["red-link"],expectPageIdToEqual:"2"});
		expectAddressTag("[-nonexistentPageAlias]", {expectTextToEqual:"nonexistentPageAlias",expectClassToContain:["red-link"],expectPageIdToEqual:"nonexistentPageAlias"});
	});
});

