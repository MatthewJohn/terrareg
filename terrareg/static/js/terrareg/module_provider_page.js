/*
 * Setup router and call setup page depending on the page/module type
 */
function renderPage() {
    const router = new Navigo("/");

    const baseRoute = "/modules/:namespace/:module/:provider";

    // Base module provider route
    router.on(baseRoute, function ({ data }) {
        setupBasePage(data);
        setupRootModulePage(data);
    });
    // Base module version route
    router.on(baseRoute + "/:version", function ({ data }) {
        setupBasePage(data);
        setupRootModulePage(data);
    });
    // Submodule route
    router.on(baseRoute + "/:version/submodule/(.*)", ({ data }) => {
        setupBasePage(data);
        setupSubmodulePage(data);
    });

    router.resolve();
}

/*
 * Generate terrareg module ID based on data from URL route
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

    id += `/${data.module}`;
    if (stopAt.module) {
        return id;
    }

    id += `/${data.provider}`;
    if (stopAt.provider) {
        return id;
    }

    if (! data.version) {
        return id;
    }
    id += `/${data.version}`;
    if (stopAt.version) {
        return id;
    }

    return id;
}

/*
 * Set the provider logo, if available and add TOS
 * to bottom of the page
 *
 * @param moduleDetails Terrareg module details
 */
async function setProviderLogo(moduleDetails) {
    let providerLogos = await getProviderLogos();

    // Check if namespace has a logo
    if (providerLogos[moduleDetails.provider] !== undefined) {
        let logoDetails = providerLogos[moduleDetails.provider];

        let logoLink = $("#provider-logo-link");
        logoLink.attr("href", logoDetails.link);

        let logoImg = $("#provider-logo-img");
        logoImg.attr("src", logoDetails.source);
        logoImg.attr("alt", logoDetails.alt);

        addProviderLogoTos(moduleDetails.provider);

        logoLink.show();
    }
}

/*
 * Add options to version selection dropdown.
 * Show currently selected module and latest version
 *
 * @param moduleDetails Terrareg module details
 */
function createVersionDetails(moduleDetails) {
    let versionSelection = $("#version-select");

    let currentVersionFound = false;
    moduleDetails.versions.forEach((version, itx) => {
        let versionOption = $("<option></option>");

        // Set value of option to view URL of module version
        versionOption.val(`/modules/${moduleDetails.namespace}/${moduleDetails.name}/${moduleDetails.provider}/${version}`);

        let versionText = version;
        // Add '(latest)' suffix to the first (latest) version
        if (itx == 0) {
            versionText += " (latest)";
        }
        versionOption.text(versionText);

        // Set version that matches current module to selected
        if (version == moduleDetails.version) {
            versionOption.attr("selected", "");
            currentVersionFound = true;
        }

        versionSelection.append(versionOption);
    });

    // If current version has not been found, add fake version to drop-down
    if (currentVersionFound == false) {
        let versionOption = $("<option></option>");
        versionOption.text(`${moduleDetails.version} (unpublished)`);
        versionOption.attr("selected", "");
        versionSelection.append(versionOption);

        // Add warning to page about unpublished version
        $("#unpublished-warning").show();
    }
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
 * @param moduleDetails Terrareg module details
 */
function setModuleTitle(moduleDetails) {
    $("#module-title").text(moduleDetails.name);
}

/*
 * Set the module provider text
 *
 * @param moduleDetails Terrareg module details
 */
function setModuleProvider(moduleDetails) {
    $("#module-provider").text(`Provider: ${moduleDetails.provider}`);
}

/*
 * Set the module description text
 *
 * @param moduleDetails Terrareg module details
 */
function setModuleDescription(moduleDetails) {
    $("#module-description").text(moduleDetails.description);
}

function showModuleDetailsBody() {
    $("#module-details-body").show();
}

/*
 * Set warning on page that there are no available versions of the module
 *
 * @param moduleDetails Terrareg module details
 */
function showNoAvailableVersions() {
    $("#no-version-available").show();
}

/*
 * Set text for 'published at' and link to parent namespace
 *
 * @param moduleDetails Terrareg module details
 */
function setPublishedAt(moduleDetails) {
    let publishedAtDiv = $("#published-at");

    let namespaceLinkDiv = $("<a></a>");
    namespaceLinkDiv.attr("href", `/modules/${moduleDetails.namespace}`);
    namespaceLinkDiv.text(moduleDetails.namespace);

    publishedAtDiv.append(`Published ${moduleDetails.published_at_display} by `);
    publishedAtDiv.append(namespaceLinkDiv);
}

/*
 * Set text for 'owner' of module, if this has been provided
 *
 * @param moduleDetails Terrareg module details
 */
function setOwner(moduleDetails) {
    if (moduleDetails.owner) {
        $("#module-owner").text(`Module managed by ${moduleDetails.owner}`);
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
 * Populate submodule selection options
 *
 * @param moduleDetails Terrareg module details
 */
function populateSubmoduleSelect(moduleDetails, currentSubmodulePath = undefined) {
    $.get(`/v1/terrareg/modules/${moduleDetails.id}/submodules`, (data) => {
        if (data.length) {
            $("#submodule-select-container").show();
        }

        let submoduleSelect = $("#submodule-select");

        data.forEach((submodule) => {
            let selectOption = $("<option></option>");
            selectOption.val(submodule.href);
            selectOption.text(submodule.path);

            // Check if current submodule path matches the item and mark as selected
            if (currentSubmodulePath !== undefined && submodule.path == currentSubmodulePath) {
                selectOption.attr('selected', '');
            }

            submoduleSelect.append(selectOption);
        });
    });
}

/*
 * Populate example selection options
 *
 * @param moduleDetails Terrareg module details
 */
function populateExampleSelect(moduleDetails) {
    $.get(`/v1/terrareg/modules/${moduleDetails.id}/examples`, (data) => {
        if (data.length) {
            $("#example-select-container").show();
        }

        let exampleSelect = $("#example-select");

        data.forEach((example) => {
            let selectOption = $("<option></option>");
            selectOption.val(example.href);
            selectOption.text(example.path);
            exampleSelect.append(selectOption);
        });
    });
}

/*
 * Handle submodule/example select onChange event.
 * Redirect user to selected submodule/example
 *
 * @param event Target event
 */
function onSubmoduleSelectChange(event) {
    let target_obj = $(event.target)[0].selectedOptions[0];
    let url = target_obj.value;

    // Ignore selection of 'Select modules' item
    if (url == "none") {
        return;
    }
    // Navigate page to submodule
    window.location.href = url;
}

/*
 * Populate usage example
 *
 * @param moduleDetails Terrareg module details
 */
async function populateTerraformUsageExample(moduleDetails, additionalPath = undefined) {
    let config = await getConfig();

    // Update labels for example analytics token and phrase
    $("#usage-example-analytics-token").text(config.EXAMPLE_ANALYTICS_TOKEN);
    $("#usage-example-analytics-token-phrase").text(config.ANALYTICS_TOKEN_PHRASE);

    // Add example terraform call to source section
    $("#usage-example-terraform").text(
        `module "${moduleDetails.name}" {
source  = "${window.location.hostname}/${config.EXAMPLE_ANALYTICS_TOKEN}__${moduleDetails.module_provider_id}${additionalPath ? "//" + additionalPath : ""}"
version = "${moduleDetails.terraform_example_version_string}"

# Provide variables here
}`
    );
}

/*
 * Populate downloads summary
 *
 * @param moduleDetails Terrareg module details
 */
function populateDownloadSummary(moduleDetails) {
    $.get(`/v1/modules/${moduleDetails.module_provider_id}/downloads/summary`, function (data, status) {
        Object.keys(data.data.attributes).forEach((key) => {
            $(`#downloads-${key}`).html(data.data.attributes[key]);
        });
    });

    // Show download container
    $('#module-download-stats-container').show();
}

/*
 * Populate analytics table
 *
 * @param moduleDetails Terrareg module details
 */
function populateAnalyticsTable(moduleDetails) {
    $.get(`/v1/terrareg/analytics/${moduleDetails.module_provider_id}/token_versions`, (data, status) => {
        Object.keys(data).forEach((token) => {
            $("#analyticsVersionByTokenTable").append(`
                <tr>
                    <td>
                        ${token}
                    </td>
                    <td>
                        ${data[token].module_version}
                    </td>
                    <td>
                        ${data[token].terraform_version}
                    </td>
                    <td>
                        ${data[token].environment}
                    </td>
                </tr>
            `);
        });
        $("#analytics-table").DataTable();
    });

    // Show tab link
    $("#module-tab-link-analytics").show();
}

/*
 * Populate README tab content and tab link, if README is available
 *
 * @param moduleDetails Terrareg module details
 */
function populateReadmeContent(readmeContent) {
    // If no README is defined, exit early
    if (!readmeContent) {
        return;
    }

    let readmeContentDiv = $("#module-tab-readme");

    // Populate README conrtent
    readmeContentDiv.append(readmeContent);

    // Add 'table' class to all tables in README
    readmeContentDiv.find("table").addClass("table");

    // Replace size of headers
    readmeContentDiv.find("h1").addClass("subtitle").addClass("is-3");
    readmeContentDiv.find("h2").addClass("subtitle").addClass("is-4");
    readmeContentDiv.find("h3").addClass("subtitle").addClass("is-5");
    readmeContentDiv.find("h4").addClass("subtitle").addClass("is-6");

    // Show README tab button
    $("#module-tab-link-readme").show();
}

/*
 * Handle tab button selection.
 *
 * @param tabName Name of tab to switch to
 * @param redirect Whether to add tab anchor to page URL
 */
function selectModuleTab(tabName, redirect) {
    if (redirect !== false) {
        // Set URL anchor to selected tag
        window.location.hash = "#" + tabName;
    }

    let tabContentId = "module-tab-" + tabName;
    let tabLinkId = "module-tab-link-" + tabName;
    let i, tabContent, tabLinks;

    // Hide content of all tabs
    tabContent = document.getElementsByClassName("module-tabs");
    for (i = 0; i < tabContent.length; i++) {
        tabContent[i].style.display = "none";
    }

    // Remove 'active' from all tab links
    tabLinks = document.getElementsByClassName("module-tab-link");
    for (i = 0; i < tabLinks.length; i++) {
        tabLinks[i].className = tabLinks[i].className.replace(" is-active", "");
    }
    // Show content of current tab and mark current link as active.
    document.getElementById(tabContentId).style.display = "block";
    document.getElementById(tabLinkId).className += " is-active";
}

/*
 * Handle example file selection from example file tab
 *
 * @param filePath Example file path to load
 */
function selectExampleFile(event) {
    let selectedItem = $(event.target);
    let filePath = selectedItem.data("path");
    let moduleId = selectedItem.data("module-id");

    // Disable 'is-active' flag on all files
    $("#example-file-list-nav")
        .find("a")
        .each((itx, item) => {
            let a_obj = $(item);
            if (a_obj.data("path") == filePath) {
                item.className = "panel-block is-active";
            } else {
                item.className = "panel-block";
            }
        });
    $.ajax({
        type: "GET",
        url: `/v1/terrareg/modules/${moduleId}/example/file/${filePath}`,
        success: function (data) {
            $("#example-file-content").html(data);
        },
    });
}

/*
 * Load list of example files and populate example file nav
 *
 * @param moduleDetails Terrareg module details
 * @param exampleDetails Terrareg example details
 */

function loadExampleFileList(moduleDetails, exampleDetails) {
    $.ajax({
        type: "GET",
        url: `/v1/terrareg/modules/${moduleDetails.id}/example/filelist/${exampleDetails.path}`,
        success: function (data) {
            let firstFileSelected = false;
            data.forEach((exampleFile) => {
                $("#example-file-list-nav").append(`
                    <a class="panel-block" data-module-id="${moduleDetails.id}" data-path="${exampleFile.path}" onClick="selectExampleFile(event)">
                        <span class="panel-icon">
                            <i class="fa-solid fa-file-code" aria-hidden="true"></i>
                        </span>
                        ${example_file.filename}
                    </a>
                `);

                if (firstFileSelected == false) {
                    firstFileSelected = true;
                    selectExampleFile(exampleFile.path);
                }
            });
        },
    });
}

/*
 * Populate tabs for inputs
 *
 * @param moduleDetails Terrareg module details
 */
function populateTerrarformInputs(moduleDetails) {
    let inputTab = $("#module-tab-inputs");
    let inputTabTbody = inputTab.find("tbody");
    moduleDetails.root.inputs.forEach((input) => {
        let inputRow = $("<tr></tr>");

        let nameTd = $("<td></td>");
        nameTd.text(input.name);
        inputRow.append(nameTd);

        let descriptionTd = $("<td></td>");
        descriptionTd.text(input.description);
        inputRow.append(descriptionTd);

        let typeTd = $("<td></td>");
        typeTd.text(input.type);
        inputRow.append(typeTd);

        let defaultTd = $("<td></td>");
        defaultTd.text(input.required ? "Required" : JSON.stringify(input.default));
        inputRow.append(defaultTd);

        inputTabTbody.append(inputRow);
    });
}

/*
 * Populate tabs for outputs
 *
 * @param moduleDetails Terrareg module details
 */
function populateTerrarformOutputs(moduleDetails) {
    let outputTab = $("#module-tab-outputs");
    let outputTabTbody = outputTab.find("tbody");
    moduleDetails.root.outputs.forEach((output) => {
        let outputRow = $("<tr></tr>");

        let nameTd = $("<td></td>");
        nameTd.text(output.name);
        outputRow.append(nameTd);

        let descriptionTd = $("<td></td>");
        descriptionTd.text(output.description);
        outputRow.append(descriptionTd);

        outputTabTbody.append(outputRow);
    });
}

/*
 * Populate tabs for providers
 *
 * @param moduleDetails Terrareg module details
 */
function populateTerrarformProviders(moduleDetails) {
    let providerTab = $("#module-tab-providers");
    let providerTabTbody = providerTab.find("tbody");
    moduleDetails.root.provider_dependencies.forEach((provider) => {
        let providerRow = $("<tr></tr>");

        let nameTd = $("<td></td>");
        nameTd.text(provider.name);
        providerRow.append(nameTd);

        let namespaceTd = $("<td></td>");
        namespaceTd.text(provider.namespace);
        providerRow.append(namespaceTd);

        let sourceTd = $("<td></td>");
        sourceTd.text(provider.source);
        providerRow.append(sourceTd);

        let versionTd = $("<td></td>");
        versionTd.text(provider.version);
        providerRow.append(versionTd);

        providerTabTbody.append(providerRow);
    });
}

/*
 * Populate tabs for resources
 *
 * @param moduleDetails Terrareg module details
 */
function populateTerrarformResources(moduleDetails) {
    let resourceTab = $("#module-tab-resources");
    let resourceTabTbody = resourceTab.find("tbody");
    moduleDetails.root.resources.forEach((resource) => {
        let resourceRow = $("<tr></tr>");

        let typeTd = $("<td></td>");
        typeTd.text(resource.type);
        resourceRow.append(typeTd);

        let nameTd = $("<td></td>");
        nameTd.text(resource.name);
        resourceRow.append(nameTd);

        let providerTd = $("<td></td>");
        providerTd.text(resource.provider);
        resourceRow.append(providerTd);

        let sourceTd = $("<td></td>");
        sourceTd.text(resource.source);
        resourceRow.append(sourceTd);

        let modeTd = $("<td></td>");
        modeTd.text(resource.mode);
        resourceRow.append(modeTd);

        let versionTd = $("<td></td>");
        versionTd.text(resource.version);
        resourceRow.append(versionTd);

        let descriptionTd = $("<td></td>");
        descriptionTd.text(resource.description);
        resourceRow.append(descriptionTd);

        resourceTabTbody.append(resourceRow);
    });
}

/*
 * Index new module provider version
 */
function indexModuleVersion(moduleDetails) {
    let moduleVersionToIndex = $("#indexModuleVersion").val();
    $.post(`/v1/terrareg/modules/${moduleDetails.module_provider_id}/${moduleVersionToIndex}/import`)
        .done(() => {
            $("#index-version-success").html("Successfully indexed version");
            $("#index-version-success").css("display", "block");
            $("#index-version-error").css("display", "none");
            if ($("#indexModuleVersionPublish").is(":checked")) {
                $.post(`/v1/terrareg/modules/${moduleDetails.module_provider_id}/${moduleVersionToIndex}/publish`)
                    .done(() => {
                        $("#index-version-success").html("Successfully indexed and published version.");
                    })
                    .fail((res) => {
                        if (res.responseJSON && res.responseJSON.message) {
                            $("#index-version-error").html(res.responseJSON.message);
                        } else {
                            $("#index-version-error").html("An unexpected error occurred when publishing module version");
                        }
                        $("#index-version-error").css("display", "block");
                    });
            }
        })
        .fail((res) => {
            if (res.responseJSON && res.responseJSON.message) {
                $("#index-version-error").html(res.responseJSON.message);
            } else {
                $("#index-version-error").html("An unexpected error occurred when indexing module");
            }
            $("#index-version-error").css("display", "block");
        });
}

async function setupIntegrations(moduleDetails) {
    let config = await getConfig();
    if (config.ALLOW_MODULE_HOSTING) {
        $("#module-integrations-upload-container").show();
    }
    if (!config.PUBLISH_API_KEYS_ENABLED) {
        $("#integrations-index-module-version-publish").show();
    }

    // Setup callback method for indexing a module version
    $("#integration-index-version-button").bind("click", () => {
        indexModuleVersion(moduleDetails);
        return false;
    });

    $.get(`/v1/terrareg/modules/${moduleDetails.module_provider_id}/integrations`, (integrations) => {
        let integrationsTable = $("#integrations-table");
        integrations.forEach((integration) => {
            // Create tr for integration
            let tr = $("<tr></tr>");

            // Create td for description and put into row
            let description = $("<td></td>");
            description.text(integration.description);
            if (integration.coming_soon) {
                description.append(" <b>(Coming soon)</b>");
            }
            tr.append(description);

            // Create TD for integration details
            let contentTd = $("<td></td>");
            let codeBlock = $("<code></code>");

            let urlCodeContent = "";
            // If method was provided, prepend the URL with the method
            if (integration.method) {
                urlCodeContent += integration.method + " ";
            }
            // Add URL to code block
            urlCodeContent += integration.url;
            codeBlock.text(urlCodeContent);

            contentTd.append(codeBlock);

            // Add notes to TD on new line, if present.
            if (integration.notes) {
                contentTd.append("<br />" + integration.notes);
            }

            tr.append(contentTd);

            integrationsTable.append(tr);
        });
    });

    // Show integrations tab link
    $("#module-tab-link-integrations").show();
}

/*
 * Setup common elements of the page, shared between all types
 * of pages
 *
 * @param data Data from router
 */
async function setupBasePage(data) {
    createBreadcrumbs(data);

    let id = getCurrentObjectId(data);

    let moduleDetails = await getModuleDetails(id);

    // If module does not exist, set warning and exit
    if (moduleDetails == null) {
        showNoAvailableVersions();
        return;
    }

    showModuleDetailsBody();
    setProviderLogo(moduleDetails);

    createVersionDetails(moduleDetails);

    setModuleTitle(moduleDetails);
    setModuleProvider(moduleDetails);
    setModuleDescription(moduleDetails);
    setPublishedAt(moduleDetails);
    setOwner(moduleDetails);
    setSourceUrl(moduleDetails.display_source_url);

    addModuleLabels(moduleDetails, $("#module-title"));
}
/*
 * Setup elements of page used by root module pages.
 *
 * @param data Data from router
 */
async function setupRootModulePage(data) {
    let id = getCurrentObjectId(data);

    let moduleDetails = await getModuleDetails(id);

    populateSubmoduleSelect(moduleDetails);
    populateExampleSelect(moduleDetails);
    populateTerraformUsageExample(moduleDetails);
    populateDownloadSummary(moduleDetails);

    $.get(`/v1/terrareg/modules/${moduleDetails.id}/readme_html`, (readmeContent) => {
        populateReadmeContent(readmeContent);
    });

    populateTerrarformInputs(moduleDetails);
    populateTerrarformOutputs(moduleDetails);
    populateTerrarformProviders(moduleDetails);
    populateTerrarformResources(moduleDetails);

    populateAnalyticsTable(moduleDetails);
    setupIntegrations(moduleDetails);
}

/*
 * Setup elements of page for submodule
 *
 * @param data Data from router
 */
async function setupSubmodulePage(data) {
    let moduleVersionId = getCurrentObjectId(data);
    let moduleDetails = await getModuleDetails(moduleVersionId);

    let submodulePath = data.undefined;
    let submoduleDetails = await getSubmoduleDetails(moduleDetails.id, submodulePath);

    populateSubmoduleSelect(moduleDetails, submodulePath);
    populateExampleSelect(moduleDetails);

    populateTerraformUsageExample(moduleDetails, submodulePath);

    $.get(`/v1/terrareg/modules/${moduleDetails.id}/submodules/readme_html/${submodulePath}`, (readmeContent) => {
        populateReadmeContent(readmeContent);
    });

    populateTerrarformInputs(moduleDetails);
    populateTerrarformOutputs(moduleDetails);
    populateTerrarformProviders(moduleDetails);
    populateTerrarformResources(moduleDetails);
}


function createBreadcrumbs(data) {
    let breadcrumbs = ["Modules", data.namespace, data.module, data.provider];
    if (data.version) {
        breadcrumbs.push(data.version);
    }

    let breadcrumbUl = $("#breadcrumb-ul");
    let currentLink = "";
    breadcrumbs.forEach((breadcrumbName, itx) => {
        // Create UL item for breadcrumb
        let breadcrumbLiObject = $("<li></li>");

        // Create link to current breadcrumb item
        currentLink += `/${breadcrumbName}`;

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
