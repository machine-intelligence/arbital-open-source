'use strict';

// urlService handles working with URLs
app.service('urlService', function($http, $location, $rootScope) {
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
		if (isLive()) {
			return 'https://' + subdomain + this.getTopLevelDomain();
		} else {
			return 'http://' + subdomain + this.getTopLevelDomain();
		}
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
			var domainUrl = this.getDomainUrl();
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
});
