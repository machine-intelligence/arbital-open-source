'use strict';

// arb.urlService handles working with URLs
app.service('urlService', function($http, $location, $rootScope, stateService) {
	var that = this;

	// This will be set to true before loading content for a second page
	this.hasLoadedFirstPage = false;

	// Map of URL patterns to handlers
	this.urlRules = [];
	// Add a rule to handle URL changes
	// urlPattern - follows Angular ngRoute pattern rules
	this.addUrlHandler = function(urlPattern, rule) {
		var sections = urlPattern.split('/');
		// Match path from the beginning
		var builder = ['^'];
		var parameters = [];
		for (var n = 0; n < sections.length; n++) {
			var section = sections[n];
			if (section == 0) {
				// Ignore empty section
			} else if (section[0] == ':') {
				if (section.endsWith('?')) {
					// Optional parameter capture
					parameters.push(section.substring(1, section.length - 1));
					builder.push('(?:\\/([^\\/]+))?');
				} else {
					// Parameter capture
					parameters.push(section.substring(1));
					builder.push('\\/([^\\/]+)');
				}
			} else {
				// Match name
				builder.push('\\/' + section);
			}
		}
		// Optional trailing slash, optional query or fragment, match to end of path
		builder.push('\\/?(?:[\\?\\#].*)?$');
		rule.urlPattern = new RegExp(builder.join(''));
		rule.parameters = parameters;
		that.urlRules.push(rule);
	};

	// Construct a part of the URL with id and alias if id!=alias, otherwise just id
	this.getBaseUrl = function(base, id, alias) {
		return '/' + base + '/' + id + (alias === id ? '' : '/' + alias) + '/';
	};

	// Return the top level domain.
	this.getTopLevelDomain = function() {
		if (isLive()) {
			return 'arbital.com';
		} else {
			return 'localhost:8012';
		}
	};

	// Get a domain url (with optional subdomain)
	this.getDomainUrl = function(subdomain) {
		if (subdomain) {
			subdomain += '.';
		} else {
			subdomain = '';
		}
		subdomain = subdomain.toLowerCase();
		if (isLive()) {
			return 'https://' + subdomain + this.getTopLevelDomain();
		}
		return 'http://' + subdomain + this.getTopLevelDomain();
	};

	this.getCurrentDomainUrl = function() {
		return window.location.origin;
	};

	// Make sure the URL path is in the given canonical form, otherwise silently change
	// the URL, preserving the search() params.
	this.ensureCanonPath = function(canonPath) {
		var hash = $location.hash();
		var search = $location.search();
		this.goToUrl(canonPath, true);
		$location.hash(hash);
		for (var k in search) {
			$location.search(k, search[k]);
		}
	};

	// Go to the given url. If there is a domain switch, we refresh the page.
	this.goToUrl = function(url, replace) {
		var differentHost = false;
		if (url.indexOf('http') === 0) {
			var domainUrl = this.getCurrentDomainUrl();
			if (url.indexOf(domainUrl) !== 0) {
				differentHost = true;
			} else {
				url = url.slice(domainUrl.length);
			}
		}
		if (differentHost) {
			window.location.href = url;
		} else {
			if (replace) $location.replace();
			$location.url(url);
		}
	};

	// Returns the url for the given page.
	// options {
	//	 permalink: if true, we'll include page's id, otherwise, we'll use alias
	//	 useEditMap: if true, use edit map to retrieve info for this page
	//	 noHost: if true, don't add the host part of the URL
	//	 markId: if set, select the given mark on the page
	//	 discussionHash: if true, jump to the discussion part of the page
	//	 answersHash: if true, jump to the answers part of the page
	// }
	this.getPageUrl = function(pageId, options) {
		var options = options || {};
		var url = '/p/' + pageId + '/';
		var page = stateService.getPageFromSomeMap(pageId, options.useEditMap);

		if (page) {
			var pageId = page.pageId;
			var pageAlias = page.alias;
			// Make sure the page's alias is scoped to its group
			if (page.seeGroupId && page.pageId != page.alias) {
				var groupAlias = stateService.pageMap[page.seeGroupId].alias;
				if (pageAlias.indexOf('.') == -1) {
					pageAlias = groupAlias + '.' + pageAlias;
				}
			}

			url = that.getBaseUrl('p', options.permalink ? pageId : pageAlias, pageAlias);
			if (options.permalink) {
				url += '?l=' + pageId;
			}

			// Check page's type to see if we need a special url
			if (page.isLens()) {
				for (var n = 0; n < page.parentIds.length; n++) {
					var parent = stateService.pageMap[page.parentIds[n]];
					if (parent) {
						url = that.getBaseUrl('p', options.permalink ? parent.pageId : parent.alias, parent.alias);
						url += '?l=' + pageId;
						break;
					}
				}
			} else if (page.isComment()) {
				var parent = page.getCommentParentPage();
				if (parent) {
					url = that.getPageUrl(parent.pageId, {permalink: options.permalink, noHost: true});
					url += '#subpage-' + pageId;
				}
			}

			// Check if we should set the domain
			if (page.seeGroupId != stateService.privateGroupId) {
				if (page.seeGroupId !== '') {
					url = that.getDomainUrl(stateService.pageMap[page.seeGroupId].alias) + url;
				} else {
					url = that.getDomainUrl() + url;
				}
			}

			// Add markId argument
			if (options.markId) {
				url += url.indexOf('?') < 0 ? '?' : '&';
				url += 'markId=' + options.markId;
			}
		}
		if (url.indexOf('#') < 0) {
			if (options.discussionHash) {
				url += '#discussion';
			} else if (options.answersHash) {
				url += '#answers';
			}
		}
		var urlAlreadyHasDomain = url.length > 4 && url.substring(0,4) == 'http';
		if (!urlAlreadyHasDomain && !options.noHost) {
			url = that.getDomainUrl() + url;
		}
		return url;
	};

	// Get url to edit the given page.
	// options {
	//	 markId: if set, resolve the given mark when publishing the page and show it
	//	 parentId: if set, this will provide a quick option for adding the parent in editor
	// }
	this.getEditPageUrl = function(pageId, options) {
		options = options || {};
		var url = '';
		if (pageId in stateService.pageMap) {
			url = that.getBaseUrl('edit', pageId, stateService.pageMap[pageId].alias);
		} else {
			url = '/edit/' + pageId + '/';
		}
		// Add markId argument
		if (options.markId) {
			url += url.indexOf('?') < 0 ? '?' : '&';
			url += 'markId=' + options.markId;
		}
		if (options.parentId) {
			url += url.indexOf('?') < 0 ? '?' : '&';
			url += 'parentId=' + options.parentId;
		}
		url = that.getDomainUrl() + url;
		return url;
	};

	// Get url to create a new page.
	// options {
	//	 parentId: if set, there will be a quick option to add this page as a parent
	// }
	this.getNewPageUrl = function(options) {
		options = options || {};
		var url = '/edit/';
		if (options.parentId) {
			url += '?parentId=' + options.parentId;
		}
		url = that.getDomainUrl() + url;
		return url;
	};
});
