
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
                  200: (data) => {resolve(data)},
                  401: () => {resolve(false)}
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
