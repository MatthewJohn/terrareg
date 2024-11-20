
terraregConfigPromiseSingleton = undefined;
async function getConfig() {
    // Create promise if it hasn't already been defined
    if (terraregConfigPromiseSingleton === undefined) {
        terraregConfigPromiseSingleton = new Promise((resolve, reject) => {
            // Perform request to obtain the config
            $.ajax({
                type: "GET",
                url: "/v1/terrareg/config",
                success: function (data) {
                    resolve(data);
                }
            });
        });
    }
    return terraregConfigPromiseSingleton;
}

terraregNamespacesPromiseSingleton = undefined;
async function getNamespaces() {
    // Create promise if it hasn't already been defined
    if (terraregNamespacesPromiseSingleton === undefined) {
        terraregNamespacesPromiseSingleton = new Promise((resolve, reject) => {
            // Perform request to obtain the config
            $.ajax({
                type: "GET",
                url: "/v1/terrareg/namespaces",
                success: function (data) {
                    resolve(data);
                }
            });
        });
    }
    return terraregNamespacesPromiseSingleton;
}

terraregNamespaceDetailsPromiseSingleton = {};
async function getNamespaceDetails(namespace) {
    // Create promise if it hasn't already been defined
    if (terraregNamespaceDetailsPromiseSingleton[namespace] === undefined) {
        terraregNamespaceDetailsPromiseSingleton[namespace] = new Promise((resolve, reject) => {
            // Perform request to obtain the config
            $.ajax({
                type: "GET",
                url: `/v1/terrareg/namespaces/${namespace}`,
                statusCode: {
                    200: (data) => {resolve(data)},
                    401: () => {resolve(false)},
                    404: () => {resolve(null)}
                }
            });
        });
    }
    return terraregNamespaceDetailsPromiseSingleton[namespace];
}

async function checkInitialSetup() {
    let namespaces = await getNamespaces();
    if (! namespaces.length) {
        window.location.href = '/initial-setup'
    }
}

terraregProviderLogosPromiseSingleton = undefined;
async function getProviderLogos() {
    // Create promise if it hasn't already been defined
    if (terraregProviderLogosPromiseSingleton === undefined) {
        terraregProviderLogosPromiseSingleton = new Promise((resolve, reject) => {
            // Perform request to obtain provider logos
            $.ajax({
                type: "GET",
                url: "/v1/terrareg/provider_logos",
                success: function (data) {
                    resolve(data);
                }
            });
        });
    }
    return terraregProviderLogosPromiseSingleton;
}

terraregIsLoggedInPromiseSingleton = undefined;
async function isLoggedIn() {
    if (terraregIsLoggedInPromiseSingleton === undefined) {
        terraregIsLoggedInPromiseSingleton = new Promise((resolve, reject) => {
            $.ajax({
                url: '/v1/terrareg/auth/admin/is_authenticated',
                statusCode: {
                  200: (data) => {resolve(data)}
                }
            });
        })
    }

    return terraregIsLoggedInPromiseSingleton;
}

/*
 * Convert URL path to full URL, based on current
 * protocol and domain
 *
 * @param url Absolute URL path
 */
function pathToUrl(urlPath) {
    let fullUrl = `${window.location.protocol}//${window.location.hostname}`;
    // Check if running on non-standard port
    if (window.location.port &&
            !((window.location.protocol == 'https:' && window.location.port == 443) ||
              (window.location.protocol == 'http:' && window.location.port == 80))) {
        fullUrl += `:${window.location.port}`;
    }
    fullUrl += urlPath;
    return fullUrl;
}

/*
 * Convert failed ajax response to error message
 *
 * @param response The ajax response object
 */
function failedResponseToErrorString(response) {
    if (response.status == 401) {
        return 'You must be logged in to perform this action.<br />If you were previously logged in, please re-authentication and try again.';
    } else if (response.status == 403) {
        return 'You do not have permission to peform this action.'
    } else if (response.responseJSON && response.responseJSON.message) {
        return response.responseJSON.message
    } else {
        return 'An unexpected error occurred';
    }
}

/*
 * Obtain object of URL get parameters
 */
function getUrlParams() {
    const urlSearchParams = new URLSearchParams(window.location.search);
    return Object.fromEntries(urlSearchParams.entries());;
}

/*
 * Convert contents of HtML entity that has been imported from external HTML (i.e. converted markdown)
 */
function convertImportedHtml(element, forceHcl=false) {
    // Add 'table' class to all tables in README
    element.find("table").addClass("table");

    // Replace size of headers
    element.find("h1").addClass("subtitle").addClass("is-3");
    element.find("h2").addClass("subtitle").addClass("is-4");
    element.find("h3").addClass("subtitle").addClass("is-5");
    element.find("h4").addClass("subtitle").addClass("is-6");

    for (let codeDiv of element.find("code")) {
        if (forceHcl) {
            $(codeDiv).addClass("language-hcl");
        }
        // If code is within pre block, perform syntax highlighting.
        if (codeDiv.parentElement.nodeName.toLowerCase() == "pre") {
            window.Prism.highlightElement(codeDiv);
        } else {
            // Otherwise, removall all "language" classes
            for (let className of codeDiv.className.split(/\s+/)) {
                if (className.indexOf('language-') !== -1) {
                    $(codeDiv).removeClass(className)
                }
            }
        }
    }
}

/*
 * Populates version content from data from version endpoint
 */
function populateFooterVersionText() {
    $.get('/v1/terrareg/version').then((data) => {
        if (data.version) {
            $('#terrareg-version').text(` (${data.version})`);
        }
    })
}