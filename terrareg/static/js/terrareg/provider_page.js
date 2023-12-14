
const router = new Navigo("/");

/*
 * Setup router and call setup page depending on the page/module type
 */
function renderPage() {
    const baseRoute = "/providers/:namespace/:provider";

    // Base module provider route
    router.on({
        [baseRoute]: {
            as: "rootProvider",
            uses: function ({ data }) {
                setupBasePage(data);
            }
        }
    });
    // Base module version route
    router.on({
        [`${baseRoute}/:version`]: {
            as: "rootProviderVersion",
            uses: function ({ data }) {
                setupBasePage(data);
            }
        }
    });

    // Documentation urls
    router.on({
        [`${baseRoute}/:version/docs`]: {
            as: "docsOverview",
            uses: function({data}) {
                setupBasePage(data);
            }
        }
    })
    router.on({
        [`${baseRoute}/:version/docs/:documentationCategory/:documentationSlug`]: {
            as: "docsPage",
            uses: function({data}) {
                setupBasePage(data);
            }
        }
    })

    router.resolve();
}

/*
 * Generate terrareg provider ID based on data from URL route
 *
 * @param data Data object from router
 * @param stopAt Object specifiying level to stop at
 */
function getCurrentObjectId(data, stopAt = undefined) {
    if (stopAt === undefined) {
        stopAt = {};
    }

    let id = `${data.namespace}`;
    if (stopAt.namespace) {
        return id;
    }

    id += `/${data.provider}`;
    if (stopAt.provider) {
        return id;
    }

    if (!data.version || data.version == 'latest') {
        return id;
    }
    id += `/${data.version}`;
    if (stopAt.version) {
        return id;
    }

    return id;
}

/*
 * Generate dictionary of selected documentation from router data
 */
function getSelectedDocumentationDetails(data) {
    // If documentation type is not present,
    // return overview
    if (! data.documentationCategory && ! data.documentationSlug) {
        return {
            slug: 'index',
            category: 'overview'
        }
    }
    return {
        slug: data.documentationSlug,
        category: data.documentationCategory
    }
}

/*
 * Set the provider logo, if available and add TOS
 * to bottom of the page
 *
 * @param providerDetails Terrareg provider details
 */
async function setProviderLogo(providerDetails) {

    // Check if namespace has a logo
    if (providerDetails.logo_url) {

        let logoImg = $("#provider-logo-img");
        logoImg.attr("src", providerDetails.logo_url);
        logoImg.attr("alt", providerDetails.name);

        let logoLink = $("#provider-logo-link");
        logoLink.removeClass('default-hidden');
        logoImg.removeClass('default-hidden');
    }
}

/*
 * Populate version paragraph, instead of
 * version select
 *
 * @param providerDetails Terrareg provider details
 */
function populateVersionText(providerDetails) {
    let versionText = $("#version-text");
    versionText.text(`Version: ${providerDetails.version}`);
    versionText.removeClass('default-hidden');
}

/*
 * Add options to version selection dropdown.
 * Show currently selected module and latest version
 *
 * @param providerDetails Terrareg provider details
 */
function populateVersionSelect(providerDetails) {
    let versionSelection = $("#version-select");

    let currentVersionFound = false;
    let currentIsLatestVersion = false;

    providerDetails.versions.forEach((version, versionItx) => {
        let foundLatest = false;
        let versionOption = $("<option></option>");
        let isLatest = false;

        // Set value of option to view URL of module version
        versionOption.val(`/providers/${providerDetails.namespace}/${providerDetails.name}/${version}`);

        let versionText = version;
        // Add '(latest)' suffix to the first (latest) version
        if (versionItx == providerDetails.versions.length) {
            versionText += " (latest)";
            foundLatest = true;
            isLatest = true;
        }
        versionOption.text(versionText);

        // Set version that matches current module to selected
        if (providerDetails.version == version) {
            versionOption.attr("selected", "");
            currentVersionFound = true;

            // Determine if the current version is the latest version
            // (first in list of versions)
            if (isLatest) {
                currentIsLatestVersion = true;
            }
        }

        versionSelection.append(versionOption);
    });

    // If current version has not been found, add fake version to drop-down
    if (currentVersionFound == false) {
        let versionOption = $("<option></option>");
        versionOption.text(providerDetails.version);
        versionOption.attr("selected", "");
        versionSelection.append(versionOption);
    }
    if (!currentIsLatestVersion && !providerDetails.beta && providerDetails.published) {
        // Otherwise, if user is not viewing the latest version,
        // display warning
        $("#non-latest-version-warning").removeClass('default-hidden');
    }

    // Show version drop-down
    $('#details-version').removeClass('default-hidden');
}

/*
 * Handle version select onchange event.
 * Redirect user to newly version module version
 *
 * @param event Onchange event
 */
function onVersionSelectChange(event) {
    let target_obj = $(event.target)[0].selectedOptions[0];
    let url = target_obj.value;

    // Navigate page to version
    window.location.href = url;
}

/*
 * Set the module title text
 *
 * @param providerDetails Terrareg provider details
 */
function setProviderTitle(providerDetails) {
    $("#provider-title").text(providerDetails.name);
}

/*
 * Set the module description text
 *
 * @param providerDetails Terrareg provider details
 */
function setProviderDescription(providerDetails) {
    $("#provider-description").text(providerDetails.description);
}

function showProviderDetailsBody() {
    $("#provider-details-body").removeClass('default-hidden');
}

/*
 * Set warning on page that there are no available versions of the module
 *
 * @param providerDetails Terrareg provider details
 */
function showNoAvailableVersions() {
    $("#no-version-available").removeClass('default-hidden');
}

/*
 * Set text for 'published at' and link to parent namespace
 *
 * @param providerDetails Terrareg provider details
 */
function setPublishedAt(providerDetails) {
    let publishedAtDiv = $("#published-at");

    let namespaceLinkDiv = $("<a></a>");
    namespaceLinkDiv.attr("href", `/providers/${providerDetails.namespace}`);
    namespaceLinkDiv.text(providerDetails.namespace);

    var options = {weekday: 'short', year: 'numeric', month: 'short', day: 'numeric'};
    let date = new Date(providerDetails.published_at);
    publishedAtDiv.append(`Published ${date.toLocaleDateString("en-GB", options)} by `);
    publishedAtDiv.append(namespaceLinkDiv);
}

/*
 * Set text for 'owner' of module, if this has been provided
 *
 * @param providerDetails Terrareg provider details
 */
function setOwner(providerDetails) {
    if (providerDetails.owner) {
        $("#provider-owner").text(`Provider managed by ${providerDetails.owner}`);
    }
}

/*
 * Set text for 'source url' of module, if this has been provided
 *
 * @param sourceUrl Url of the source code
 */
function setSourceUrl(sourceUrl) {
    if (sourceUrl) {
        let sourceLink = $("<a></a>");
        sourceLink.text(sourceUrl);
        sourceLink.attr("href", sourceUrl);

        let sourceUrlDiv = $("#source-url");
        sourceUrlDiv.text("Source code: ");
        sourceUrlDiv.append(sourceLink);
    }
}

/*
 * Set custom links
 *
 * @param providerDetails
 */
function populateCustomLinks(providerDetails) {
    let customLinkParent = $('#custom-links');
    for (let linkDetails of providerDetails.custom_links) {
        let link = $('<a></a>');
        link.addClass('custom-link');
        link.attr('href', linkDetails.url);
        link.text(linkDetails.text);
        customLinkParent.append(link);
        customLinkParent.append('<br />');
    }
}

/*
 * Populate usage example
 *
 * @param providerDetails Terrareg provider details
 */
async function populateTerraformUsageExample(providerDetails) {
    // Add example Terraform call to source section
    $("#usage-example-terraform").text(`terraform {
  required_providers {
    ${providerDetails.name} = {
      source = "${window.location.host}/${providerDetails.namespace}/${providerDetails.name}"
      version = "${providerDetails.version}"
    }
  }
}

provider "${providerDetails.name}" {
  # Add provider configuration here
}`);

    // Perform syntax highlighting
    window.Prism.highlightElement(document.getElementById("usage-example-terraform"));

    // Show container
    $('#usage-example-container').removeClass('default-hidden');
}

/*
 * Handle clicking icon for expanding/collapsing the example
 */
function onProviderUsageExampleClick() {
    let usageExampleBody = $('#usage-example-body');
    let newIcon = '';
    let oldIcon = '';
    if (usageExampleBody.attr('class').split(/\s+/).indexOf("default-hidden") !== -1) {
        // Handle clicking expand
        usageExampleBody.removeClass('default-hidden');
        oldIcon = 'fa-angle-left'
        newIcon = 'fa-angle-down';
    } else {
        usageExampleBody.addClass('default-hidden');
        oldIcon = 'fa-angle-down'
        newIcon = 'fa-angle-left';
    }
    $('#usage-example-expand-icon').removeClass(oldIcon).addClass(newIcon);
}

/*
 * Populate downloads summary
 *
 * @param providerDetails Terrareg provider details
 */
function populateDownloadSummary(providerV2Details) {
    $.get(`/v2/providers/${providerV2Details.data.id}/downloads/summary`, function (data, status) {
        Object.keys(data.data.attributes).forEach((key) => {
            $(`#downloads-${key}`).html(data.data.attributes[key]);
        });
    });

    // Show download container
    $('#provider-download-stats-container').removeClass('default-hidden');
}

/*
 * Show warning that extraction is out of date
 */
function showOutdatedExtractionDataWarning(providerDetails) {
    if (providerDetails.provider_extraction_up_to_date === false) {
        $('#outdated-extraction-warning').removeClass('default-hidden');
    }
}

/*
 * Set HTML page title
 *
 * @param id Provider id
 */
function setPageTitle(id) {
    document.title = `${id} - Terrareg`;
}

/*
 * Redirect user to document page
 */
function redirectDocumentPage(providerDetails, slug, category) {
    window.location = router.generate(
        'docsPage',
        {namespace: providerDetails.namespace, provider: providerDetails.name, version: providerDetails.version,
         documentationCategory: category, documentationSlug: slug},
        {includeRoot: true, replaceRegexGroups: false}
    )
}

/*
 * Populate documentation menu with each available doc
 */
function populateDocumentationMenu(providerDetails) {
    let docCountByCategory = {
        resources: {},
        "data-sources": {},
        "guides": {},
        overview: {}
    }
    providerDetails.docs.sort((a, b) => a.slug < b.slug).forEach((doc) => {
        if (docCountByCategory[doc.category] !== undefined) {
            let linkName = doc.title;
            if (["resources", "data-sources"].indexOf(doc.category) !== -1) {
                linkName = `${providerDetails.name}_${doc.title}`;
            }
            let linkDiv = $(`
            <a id="doclink-${doc.category}-${doc.slug}" class="navbar-item">
                ${linkName}
            </a>`);
            linkDiv.bind('click', () => {redirectDocumentPage(providerDetails, doc.slug, doc.category)});
            docCountByCategory[doc.category][doc.title] = linkDiv;
        }
    });

    function addDocLinksToPage(parent, docs) {
        // Sort all keys by name and iterate over them, adding to list
        Object.keys(docs).sort((a, b) => {a > b}).forEach((docName) => {
            docs[docName].insertAfter(parent);
        });

        // If there are any docs, show the header
        if (Object.keys(docs).length) {
            parent.removeClass("default-hidden");
        }
    }

    addDocLinksToPage($('#provider-docs-menu-guides-header'), docCountByCategory.guides);
    addDocLinksToPage($('#provider-docs-menu-resources-header'), docCountByCategory.resources);
    addDocLinksToPage($('#provider-docs-menu-data-sources-header'), docCountByCategory["data-sources"]);

    // Bind overview button to link to overview page
    $('#doclink-overview-index').on('click', () => {
        window.location = router.generate('docsOverview', {
            namespace: providerDetails.namespace,
            provider: providerDetails.name,
            version: providerDetails.version
        });
    });
}

/*
 * Obtain the currently selected documentation
 * and display to the user
 */
function showSelectedDocument(providerVersionV2Details, selectedDocumentation) {
    if (! selectedDocumentation.slug || ! selectedDocumentation.category) {
        console.log("Invalid selected documentation");
        return;
    }
    // Query for documentation
    $.ajax({
        url: '/v2/provider-docs',
        type: "get",
        data: {
            'filter[provider-version]': providerVersionV2Details.id,
            'filter[category]': selectedDocumentation.category,
            'filter[slug]': selectedDocumentation.slug,
            'filter[language]': 'hcl',
            'page[size]': 1
        },

        success: (response) => {
            if (response.data && response.data.length) {
                obtainDocumentation(response.data[0].id)
            } else {
                showDocumentationError();
            }
        },
        failure: (xhr) => {
            showDocumentationError();
        }
    });

    // Highlight link to documentation
    let linkElement = $(`#doclink-${selectedDocumentation.category}-${selectedDocumentation.slug}`);
    if (linkElement) {
        linkElement.addClass("is-active");
    }
}

/*
 * Show warning that documentation page does not exist
 */
function showDocumentationError() {
    $('#provider-doc-content').html("<h3 class='subtitle is-3'>Error</h3>This documentation page does not exist")
}

/*
 * Obtain documentation by ID and populate documentation view
 */
function obtainDocumentation(documentationId) {
    let contentDiv = $('#provider-doc-content');
    contentDiv.html('');
    $.get(`/v2/provider-docs/${documentationId}?output=html`).then((data) => {
        if (data.data.attributes.content) {
            contentDiv.html(data.data.attributes.content);
            convertImportedHtml(contentDiv, true);
        }
    })
}

/*
 * Get redirect URL if URL does not match actual
 * provider details, meaning it's
 * obtained details for a redirected provider
 *
 * @param data Route data
 * @param providetDetails Provider details for provider
 *
 * @returns null if no redirect or string of redirect URL
 */
function getRedirectUrl(data, providerDetails) {
    // Check for any redirects by comparing
    // providerDetails and URL attributes
    if (data.namespace !== providerDetails.namespace ||
        data.provider !== providerDetails.name
    ) {
        // Generate redirect
        let currentRoutes = router.lastResolved();
        if (currentRoutes.length) {
            let currentRoute = currentRoutes[0];

            let redirectData = Object.assign({}, data);
            redirectData.namespace = providerDetails.namespace;
            redirectData.provider = providerDetails.name;
            
            // Generate new URL using current route data,
            // correcting the namespace, module and provider
            let newUrl = router.generate(
                currentRoute.route.name,
                redirectData,
                {includeRoot: true, replaceRegexGroups: true}
            );

            // Copy query string
            if (currentRoute.queryString)
                newUrl += `?${currentRoute.queryString}`;

            // Copy hash
            if (currentRoute.hashString)
                newUrl += `#${currentRoute.hashString}`;

            // Return new redirect URL
            return newUrl;
        }
    }
    return null;
}

/*
 * Setup common elements of the page, shared between all types
 * of pages
 *
 * @param data Data from router
 */
async function setupBasePage(data) {

    let id = getCurrentObjectId(data);

    let providerDetails = await getProviderDetails(id);
    let providerV2Details = await getV2ProviderDetails(`${providerDetails.namespace}/${providerDetails.name}`, "provider-versions");

    let redirectUrl = getRedirectUrl(data, providerDetails);
    if (redirectUrl) {
        window.location.href = redirectUrl;
        // Return early to stop rendering the page
        return;
    }

    createBreadcrumbs(data);

    setPageTitle(providerDetails.id);
    setProviderDescription(providerDetails);
    setPublishedAt(providerDetails);
    setOwner(providerDetails);

    // If current version is not available or there are no
    // versions, set warning and exit
    if (!providerDetails.version) {
        showNoAvailableVersions();
        return;
    }

    let providerVersionV2Details = providerV2Details.included.filter((a) => a.type === "provider-versions" && a.attributes.version == providerDetails.version);
    if (! providerVersionV2Details.length) {
        showNoAvailableVersions();
        return;
    }
    providerVersionV2Details = providerVersionV2Details[0];

    showProviderDetailsBody();
    // enableTerraregExclusiveTags();
    setProviderLogo(providerDetails);

    setProviderTitle(providerDetails);

    addProviderLabels(providerDetails, $("#provider-labels"));

    // showOutdatedExtractionDataWarning(providerDetails);
    populateVersionSelect(providerDetails);
    populateTerraformUsageExample(providerDetails);
    populateDownloadSummary(providerV2Details);
    setSourceUrl(providerDetails.source);
    // populateCustomLinks(providerDetails);

    populateDocumentationMenu(providerDetails);

    let selectedDocumentation = getSelectedDocumentationDetails(data);
    showSelectedDocument(providerVersionV2Details, selectedDocumentation);
}

async function createBreadcrumbs(data, subpath = undefined) {
    let namespaceName = data.namespace;
    let namespaceDetails = await getNamespaceDetails(namespaceName);
    if (namespaceDetails.display_name) {
        namespaceName = namespaceDetails.display_name;
    }

    let breadcrumbs = [
        ["Providers", "providers"],
        [namespaceName, data.namespace],
        [data.provider, data.provider]
    ];
    if (data.version) {
        breadcrumbs.push([data.version, data.version]);
    }

    let breadcrumbUl = $("#breadcrumb-ul");
    let currentLink = "";
    breadcrumbs.forEach((breadcrumbDetails, itx) => {
        let breadcrumbName = breadcrumbDetails[0];
        let breadcrumbUrlPart = breadcrumbDetails[1];

        // Create UL item for breadcrumb
        let breadcrumbLiObject = $("<li></li>");

        // Create link to current breadcrumb item
        currentLink += `/${breadcrumbUrlPart}`;

        // Create link for breadcrumb
        let breadcrumbLink = $(`<a></a>`);
        breadcrumbLink.attr("href", currentLink);
        breadcrumbLink.text(breadcrumbName);

        // If hitting last breadcrumb, set as active
        if (itx == breadcrumbs.length - 1) {
            breadcrumbLiObject.addClass("is-active");
        }

        breadcrumbLiObject.append(breadcrumbLink);

        // Add breadcrumb item to to list
        breadcrumbUl.append(breadcrumbLiObject);
    });
}
