
class BaseTab {
    constructor() {
        this._renderPromise = undefined;
    }
    render() {}
    async isValid() {
        let result = await this._renderPromise;
        return result;
    }
}

class ModuleDetailsTab extends BaseTab {
    constructor(moduleDetails) {
        super();
        this._moduleDetails = moduleDetails;
    }
    render() {}
}

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
    // Example route
    router.on(baseRoute + "/:version/example/(.*)", ({ data }) => {
        setupBasePage(data);
        setupExamplePage(data);
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

        logoLink.removeClass('default-hidden');
        logoImg.removeClass('default-hidden');
    }
}

/*
 * Populate version paragraph, instead of
 * version select
 *
 * @param moduleDetails Terrareg module details
 */
function populateVersionText(moduleDetails) {
    let versionText = $("#version-text");
    versionText.text(`Version: ${moduleDetails.version}`);
    versionText.removeClass('default-hidden');
}

/*
 * Add options to version selection dropdown.
 * Show currently selected module and latest version
 *
 * @param moduleDetails Terrareg module details
 */
function populateVersionSelect(moduleDetails) {
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
        $("#unpublished-warning").removeClass('default-hidden');
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
    $("#module-details-body").removeClass('default-hidden');
}

/*
 * Set warning on page that there are no available versions of the module
 *
 * @param moduleDetails Terrareg module details
 */
function showNoAvailableVersions() {
    $("#no-version-available").removeClass('default-hidden');
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
            $("#submodule-select-container").removeClass('default-hidden');
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
function populateExampleSelect(moduleDetails, currentSubmodulePath = undefined) {
    $.get(`/v1/terrareg/modules/${moduleDetails.id}/examples`, (data) => {
        if (data.length) {
            $("#example-select-container").removeClass('default-hidden');
        }

        let exampleSelect = $("#example-select");

        data.forEach((example) => {
            let selectOption = $("<option></option>");
            selectOption.val(example.href);
            selectOption.text(example.path);

            // Check if current submodule path matches the item and mark as selected
            if (currentSubmodulePath !== undefined && example.path == currentSubmodulePath) {
                selectOption.attr('selected', '');
            }

            exampleSelect.append(selectOption);
        });
    });
}

/*
 * Populate and show current submodule/example text
 *
 * @parram text Text to be shown
 */
function populateCurrentSubmodule(text) {
    let currentSubmodule = $('#current-submodule');
    currentSubmodule.text(text);
    currentSubmodule.removeClass('default-hidden');
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

    // Show container
    $('#usage-example-container').removeClass('default-hidden');
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
    $('#module-download-stats-container').removeClass('default-hidden');
}



class AnalyticsTab extends ModuleDetailsTab {
    get name() {
        return 'analytics';
    }
    async render() {
        this._renderPromise = new Promise(async (resolve) => {
            $.get(`/v1/terrareg/analytics/${this._moduleDetails.module_provider_id}/token_versions`, (data, status) => {
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
            $("#module-tab-link-analytics").removeClass('default-hidden');

            resolve(true);
        });
    }
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
    $.find('.module-tabs').forEach((div) => {
        $(div).addClass('default-hidden');
    });

    // Remove 'active' from all tab links
    $.find('.module-tab-link').forEach((tabLinkDiv) => {
        $(tabLinkDiv).removeClass('is-active');
    });

    // Show content of current tab and mark current link as active.
    $(`#${tabContentId}`).removeClass('default-hidden');
    $(`#${tabLinkId}`).addClass('is-active');
}

/*
 * Handle example file selection from example file tab
 *
 * @param filePath Example file path to load
 */
function selectExampleFile(eventTarget) {
    let selectedItem = $(eventTarget);
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
        url: `/v1/terrareg/modules/${moduleId}/examples/file/${filePath}`,
        success: function (data) {
            $("#example-file-content").html(data);
        },
    });
}

class ExampleFilesTab extends ModuleDetailsTab {
    constructor(moduleDetails, exampleDetails) {
        super(moduleDetails);
        this._exampleDetails = exampleDetails;
    }
    get name() {
        return 'example-files';
    }
    async render() {
        this._renderPromise = new Promise(async (resolve) => {
            $.get(`/v1/terrareg/modules/${this._moduleDetails.id}/examples/filelist/${this._exampleDetails.path}`, (data) => {
                if (! data.length) {
                    resolve(false);
                    return;
                }

                let firstFileSelected = false;
                data.forEach((exampleFile) => {
                    let fileLink = $(`
                        <a class="panel-block" data-module-id="${this._moduleDetails.id}" data-path="${exampleFile.path}" onClick="selectExampleFile(event.target)">
                            <span class="panel-icon">
                                <i class="fa-solid fa-file-code" aria-hidden="true"></i>
                            </span>
                            ${exampleFile.filename}
                        </a>
                    `);

                    $("#example-file-list-nav").append(fileLink);

                    if (firstFileSelected == false) {
                        firstFileSelected = true;
                        selectExampleFile(fileLink);
                    }
                });

                // Show example tab link
                $('#module-tab-link-example-files').removeClass('default-hidden');

                resolve(true);
            });
        });
    }
}

class InputsTab extends ModuleDetailsTab {
    get name() {
        return 'inputs';
    }
    async render() {
            this._renderPromise = new Promise(async (resolve) => {
            let inputTab = $("#module-tab-inputs");
            let inputTabTbody = inputTab.find("tbody");
            this._moduleDetails.inputs.forEach((input) => {
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

            // Show tab link
            $('#module-tab-link-inputs').removeClass('default-hidden');
            resolve(true);
        });
    }
}

class OutputsTab extends ModuleDetailsTab {
    get name() {
        return 'outputs';
    }
    async render() {
        this._renderPromise = new Promise(async (resolve) => {
            let outputTab = $("#module-tab-outputs");
            let outputTabTbody = outputTab.find("tbody");
            this._moduleDetails.outputs.forEach((output) => {
                let outputRow = $("<tr></tr>");

                let nameTd = $("<td></td>");
                nameTd.text(output.name);
                outputRow.append(nameTd);

                let descriptionTd = $("<td></td>");
                descriptionTd.text(output.description);
                outputRow.append(descriptionTd);

                outputTabTbody.append(outputRow);
            });
            // Show tab link
            $('#module-tab-link-outputs').removeClass('default-hidden');

            resolve(true);
        });
    }
}

class ProvidersTab extends ModuleDetailsTab {
    get name() {
        return 'providers';
    }
    async render() {
        this._renderPromise = new Promise(async (resolve) => {
            let providerTab = $("#module-tab-providers");
            let providerTabTbody = providerTab.find("tbody");
            this._moduleDetails.provider_dependencies.forEach((provider) => {
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

            // Show tab link
            $('#module-tab-link-providers').removeClass('default-hidden');
            resolve(true);
        });
    }
}

class ResourcesTab extends ModuleDetailsTab {
    get name() {
        return 'resources';
    }
    async render() {
        this._renderPromise = new Promise(async (resolve) => {
            let resourceTab = $("#module-tab-resources");
            let resourceTabTbody = resourceTab.find("tbody");
            this._moduleDetails.resources.forEach((resource) => {
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

            // Show tab link
            $('#module-tab-link-resources').removeClass('default-hidden');
            resolve(true);
        });
    }
}

/*
 * Index new module provider version
 */
function indexModuleVersion(moduleDetails) {
    let moduleVersionToIndex = $("#indexModuleVersion").val();
    $.post(`/v1/terrareg/modules/${moduleDetails.module_provider_id}/${moduleVersionToIndex}/import`)
        .done(() => {
            $("#index-version-success").html("Successfully indexed version");
            $("#index-version-success").addClass('default-hidden');
            $("#index-version-error").addClass('default-hidden');
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
                        $("#index-version-error").removeClass('default-hidden');
                    });
            }
        })
        .fail((res) => {
            if (res.responseJSON && res.responseJSON.message) {
                $("#index-version-error").html(res.responseJSON.message);
            } else {
                $("#index-version-error").html("An unexpected error occurred when indexing module");
            }
            $("#index-version-error").removeClass('default-hidden');
        });
}

class IntegrationsTab extends ModuleDetailsTab {
    get name() {
        return 'integrations';
    }
    async render() {
        this._renderPromise = new Promise(async (resolve) => {
            let config = await getConfig();
            if (config.ALLOW_MODULE_HOSTING) {
                $("#module-integrations-upload-container").removeClass('default-hidden');
            }
            if (!config.PUBLISH_API_KEYS_ENABLED) {
                $("#integrations-index-module-version-publish").removeClass('default-hidden');
            }

            // Setup callback method for indexing a module version
            $("#integration-index-version-button").bind("click", () => {
                indexModuleVersion(this._moduleDetails);
                return false;
            });

            $.get(`/v1/terrareg/modules/${this._moduleDetails.module_provider_id}/integrations`, (integrations) => {
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
            $("#module-tab-link-integrations").removeClass('default-hidden');

            resolve(true);
        });
    }
}

/*
 * Setup link to parent root module page
 *
 * @param moduleDetails Terrareg module details
 */
function enableBackToParentLink(moduleDetails) {
    let backToParentLink = $('#submodule-back-to-parent');
    backToParentLink.attr('href', `/modules/${moduleDetails.id}`);
    backToParentLink.removeClass('default-hidden');
}

/*
 * Enable 'terrareg exclusive' tags, if enabled in the config
 */
async function enableTerraregExclusiveTags() {
    let config = await getConfig();
    if (! config.DISABLE_TERRAREG_EXCLUSIVE_LABELS) {
        $.find('.terrareg-exclusive').forEach((tag) => {
            $(tag).removeClass('default-hidden');
        });
    }
}

class SettingsTab extends ModuleDetailsTab {
    get name() {
        return 'settings';
    }
    async render() {
        this._renderPromise = new Promise(async (resolve) => {

            let loggedIn = await isLoggedIn();
            // Return immediately if user is not logged in
            if (! loggedIn) {
                resolve(false);
                return;
            }

            if (this._moduleDetails.verified) {
                $('#settings-verified').attr('checked', true);
            }

            // Check if namespace is auto-verified and, if so, show message
            $.get(`/v1/terrareg/modules/${this._moduleDetails.namespace}`, (namespaceDetails) => {
                if (namespaceDetails.is_auto_verified) {
                    $('#settings-verified-auto-verified-message').removeClass('default-hidden');
                }
            });

            // Setup git providers
            let config = await getConfig();
            let gitProviderSelect = $('#settings-git-provider');

            if (config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER || config.ALLOW_CUSTOM_GIT_URL_MODULE_VERSION) {
                let customGitProvider = $('<option></option>');
                customGitProvider.val('');
                customGitProvider.text('Custom');
                if (! this._moduleDetails.git_provider_id) {
                    customGitProvider.attr('selected', '');
                }
                gitProviderSelect.append(customGitProvider);
            }

            // Obtain all git providers and add to select for settings
            $.get('/v1/terrareg/git_providers', (gitProviders) => {
                gitProviders.forEach((gitProvider) => {
                    let gitProviderOption = $('<option></option>');
                    gitProviderOption.val(gitProvider.id);
                    gitProviderOption.text(gitProvider.name);
                    if (this._moduleDetails.git_provider_id == gitProvider.id) {
                        gitProviderOption.attr('selected', '');
                    }
                    gitProviderSelect.append(gitProviderOption);
                });
            });

            $('#settings-git-tag-format').val(this._moduleDetails.git_tag_format);

            let baseUrlTemplate = $('#settings-base-url-template');
            baseUrlTemplate.attr('placeholder', `https://github.com/${this._moduleDetails.namespace}/${this._moduleDetails.name}-${this._moduleDetails.provider}`);
            baseUrlTemplate.val(this._moduleDetails.repo_base_url_template);

            let cloneUrlTemplate = $('#settings-clone-url-template');
            cloneUrlTemplate.attr('placeholder', `ssh://git@github.com/${this._moduleDetails.namespace}/${this._moduleDetails.name}-${this._moduleDetails.provider}.git`);
            cloneUrlTemplate.val(this._moduleDetails.repo_clone_url_template);

            let browseUrlTemplate = $('#settings-browse-url-template');
            browseUrlTemplate.attr('placeholder', `https://github.com/${this._moduleDetails.namespace}/${this._moduleDetails.name}-${this._moduleDetails.provider}/tree/{tag}/{path}`);
            browseUrlTemplate.val(this._moduleDetails.repo_browse_url_template);

            // Bind module provider delete button
            let moduleProviderDeleteButton = $('#module-provider-delete-button');
            moduleProviderDeleteButton.bind('click', () => {
                deleteModuleProvider(this._moduleDetails);
            });

            // Bind settings form submission with function
            $('#settings-form').submit(() => {
                updateModuleProviderSettings(this._moduleDetails);
                return false;
            });

            // Enable custom git provider inputs
            if (config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER) {
                $.find('.settings-custom-git-provider-container').forEach((inputContainer) => {
                    $(inputContainer).removeClass('default-hidden');
                });
            }

            // Show settings tab
            $('#module-tab-link-settings').removeClass('default-hidden');
            resolve(true);
        });
    }
}

function setupModuleVersionDeletionSetting(moduleDetails) {
    // Bind module version deletion with deleteModuleVersion function
    let moduleVersionDeleteButton = $('#module-version-delete-button');
    moduleVersionDeleteButton.bind('click', () => {
        deleteModuleVersion(moduleDetails);
    });
    moduleVersionDeleteButton.text(`Delete Module Version: ${moduleDetails.version}`);
    // Setup confirmation checkbox for confirm deletion
    let confirmDeleteContainer = $('#confirm-delete-module-version-div');
    confirmDeleteContainer.append(`Confirm deletion of module version ${moduleDetails.version}: `);
    let confirmCheckbox = $('<input />');
    confirmCheckbox.attr('autocomplete', 'off');
    confirmCheckbox.attr('type', 'checkbox');
    confirmCheckbox.attr('id', 'confirm-delete-module-version');
    confirmDeleteContainer.append(confirmCheckbox);

    // Show module version deletion settings
    $('#module-version-delete-container').removeClass('default-hidden');
}

function deleteModuleVersion(moduleDetails) {
    $('#confirm-delete-module-version-div').css('display', 'block');
    if (! $('#confirm-delete-module-version').is(':checked')) {
        return;
    }
    $.ajax({
        url: `/v1/terrareg/modules/${moduleDetails.id}/delete`,
        method: 'delete',
        data: JSON.stringify({
            csrf_token: $('#settings-csrf-token').val()
        }),
        contentType: 'application/json'
    }).done(()=> {
        // Reidrect to module provider page.
        // This will automatically redirect back
        // down the path to valid existing object.
        window.location.href = `/modules/${ moduleDetails.module_provider_id }`;
    });
    return false;
}

function deleteModuleProvider(moduleDetails) {
    $('#confirm-delete-module-provider-div').removeClass('default-hidden');
    if (! $('#confirm-delete-module-provider').is(':checked')) {
        return;
    }
    $.ajax({
        url: `/v1/terrareg/modules/${moduleDetails.module_provider_id}/delete`,
        method: 'delete',
        data: JSON.stringify({
            csrf_token: $('#settings-csrf-token').val()
        }),
        contentType: 'application/json'
    }).done(()=> {
        // Reidrect to module page.
        // This will automatically redirect back
        // down the path to valid existing object.
        window.location.href = `/modules/${moduleDetails.id}`;
    });
}

/*
 * Handle form submission for changing module provider settings
 *
 * @param ev Event of target form
 */
function updateModuleProviderSettings(moduleDetails) {
    $('#settings-status-success').addClass('default-hidden');
    $('#settings-status-error').addClass('default-hidden');
    $.post({
        url: `/v1/terrareg/modules/${moduleDetails.module_provider_id}/settings`,
        data: JSON.stringify({
            git_provider_id: $('select[id=settings-git-provider] option').filter(':selected').val(),
            repo_base_url_template: $('#settings-base-url-template').val(),
            repo_clone_url_template: $('#settings-clone-url-template').val(),
            repo_browse_url_template: $('#settings-browse-url-template').val(),
            git_tag_format: $('#settings-git-tag-format').val(),
            verified: $('#settings-verified').is(':checked'),
            csrf_token: $('#settings-csrf-token').val()
        }),
        contentType: 'application/json'
    }).done(() => {
        $('#settings-status-success').removeClass('default-hidden');
    }).fail((res) => {
        if (res.responseJSON && res.responseJSON.message) {
            $('#settings-status-error').html(res.responseJSON.message);
        } else {
            $('#settings-status-error').html('An unexpected error occurred');
        }
        $('#settings-status-error').removeClass('default-hidden');
    });

    // Return false to present default action
    return false;
}

class UsageBuilderTab extends ModuleDetailsTab {
    get name() {
        return 'usage-builder';
    }
    async render() {
        this._renderPromise = new Promise(async (resolve) => {
            let config = await getConfig();
            let usageBuilderTable = $('#usageBuilderTable');

            // Setup analaytics input row
            let analyticsTokenInputRow = $('<tr></tr>');

            let analyticsTokenName = $('<td></td>');
            analyticsTokenName.text(config.ANALYTICS_TOKEN_PHRASE);
            analyticsTokenInputRow.append(analyticsTokenName);

            let analyticsTokenDescription = $('<td></td>');
            analyticsTokenDescription.text(config.ANALYTICS_TOKEN_DESCRIPTION);
            analyticsTokenInputRow.append(analyticsTokenDescription);

            let analyticsTokenInputTd = $('<td></td>');
            let analyticsTokenInputField = $('<input />');
            analyticsTokenInputField.attr('class', 'input');
            analyticsTokenInputField.attr('id', 'usageBuilderAnalyticsToken');
            analyticsTokenInputField.attr('type', 'text');
            analyticsTokenInputField.attr('placeholder', config.EXAMPLE_ANALYTICS_TOKEN);
            analyticsTokenInputField.bind('keyup', () => {updateUsageBuilderOutput(this._moduleDetails)});
            analyticsTokenInputTd.append(analyticsTokenInputField);
            analyticsTokenInputRow.append(analyticsTokenInputTd);

            usageBuilderTable.append(analyticsTokenInputRow);

            let inputVariables = await getUsageBuilderVariables(this._moduleDetails.id);

            if (inputVariables === null || ! inputVariables.length) {
                // If there are no variables present in the usage builder,
                // then exit early
                resolve(false);
                return;
            }

            // Show tab
            $('#module-tab-link-usage-builder').removeClass('default-hidden');
            resolve(true);

            // Build input table
            inputVariables.forEach((inputVariable) => {
                let inputId = `usageBuilderInput-${inputVariable.name}`;

                let inputRow = $('<tr></tr>');
                let inputNameTd = $('<td></td>');
                inputNameTd.text(inputVariable.name);
                inputRow.append(inputNameTd);

                let additionalHelpTd = $('<td></td>');
                additionalHelpTd.text(inputVariable.additional_help ? inputVariable.additional_help : '');
                inputRow.append(additionalHelpTd);

                let valueTd = $('<td></td>');

                if (inputVariable.type == 'text') {
                    let inputDiv = $('<input />');
                    inputDiv.addClass('input');
                    inputDiv.attr('type', 'text');
                    inputDiv.attr('id', inputId);
                    inputDiv.bind('keyup', () => {updateUsageBuilderOutput(this._moduleDetails)});
                    valueTd.append(inputDiv);

                } else if (inputVariable.type == 'boolean') {

                    let inputDiv = $('<input />');
                    inputDiv.addClass('checkbox');
                    inputDiv.attr('type', 'checkbox');
                    inputDiv.attr('id', inputId);
                    inputDiv.attr('value', 'true');
                    inputDiv.bind('onchange', () => {updateUsageBuilderOutput(this._moduleDetails)});
                    valueTd.append(inputDiv);

                } else if (inputVariable.type == 'select') {
                    let inputDiv = $('<div></div>');
                    inputDiv.addClass('select');
                    let inputSelect = $('<select></select>');
                    inputSelect.attr('id', inputId);
                    inputSelect.bind('change', () => {updateUsageBuilderOutput(this._moduleDetails)});

                    inputVariable.choices.forEach((inputChoice, itx) => {
                        // If choices is list of strings, use the string as the name,
                        // otherwise, use name attribute of object.
                        let inputName = typeof inputChoice === 'string' ? inputChoice : inputChoice.name;
                        let option = $('<option></option>');
                        option.val(itx);
                        option.text(inputName);
                        inputSelect.append(option);
                    });
                    // If custom input is available, add to select
                    if (inputVariable.allow_custom) {
                        let option = $('<option></option>');
                        option.val('custom');
                        option.text('Custom Value');
                        inputSelect.append(option);
                    }
                    inputDiv.append(inputSelect);

                    valueTd.append(inputDiv);

                    // If custom input is available, add hidden input for custom input
                    let customInput = $('<input />');
                    customInput.addClass('input');
                    customInput.attr('type', 'text');
                    customInput.attr('id', `${inputId}-customValue`);
                    customInput.bind('keyup', () => {updateUsageBuilderOutput(this._moduleDetails)});
                    customInput.css('display', 'none');
                    valueTd.append(customInput);

                } else {
                    // Skip displaying other types of variables in input
                    return;
                }

                inputRow.append(valueTd);

                $('#usageBuilderTable').append(inputRow);
            });
        });
    }
}

function usageBuilderQuoteString(input) {
    // Place input value directly into double quotes.
    // Escape backslashes and then escape double quotes.
    return '"' + input.replace(/\\/g, '\\\\').replace(/"/g, '\\"') + '"';
}

async function updateUsageBuilderOutput(moduleDetails) {
    let outputTf = '';
    let additionalContent = '';

    let analytics_token = $('#usageBuilderAnalyticsToken')[0].value;

    let inputVariables = await getUsageBuilderVariables(moduleDetails.id);

    inputVariables.forEach((inputVariable) => {
        let inputId = `#usageBuilderInput-${inputVariable.name}`;
        let varInput = '';

        // Get value from
        if (inputVariable.type == 'static')
        {
            varInput = inputVariable.value;
        }
        else if (inputVariable.type == 'text')
        {
            varInput = $(inputId)[0].value;
        }
        else if (inputVariable.type == 'boolean')
        {
            varInput = $(inputId).is(':checked') ? 'true' : 'false';
        }
        else if (inputVariable.type == 'select')
        {
            let selectIndex = $(inputId)[0].value;
            let customInputId = `${inputId}-customValue`;

            // Check if custom type
            if (selectIndex == 'custom') {
                // Display custom text input
                $(customInputId)[0].style.display ='block';

                // Use value of custom input as output
                varInput = $(customInputId)[0].value
            }
            else
            {
                // Hide custom input and clear value
                $(customInputId)[0].style.display = 'none';
                $(customInputId)[0].value = '';

                // If choice is a string, add the choice name to as the value
                if (typeof inputVariable.choices[selectIndex] === 'string') {
                    varInput = inputVariable.choices[selectIndex];
                } else {
                    // Otherwise, use the attribute for the value
                    varInput = inputVariable.choices[selectIndex].value;

                    // If object has additional_content, add it to the TF output
                    if (inputVariable.choices[selectIndex].additional_content) {
                        additionalContent += inputVariable.choices[selectIndex].additional_content + '\n\n';
                    }
                }
            }
            
        }

        if (inputVariable.quote_value) {
            varInput = usageBuilderQuoteString(varInput);
        }

        outputTf += `\n  ${inputVariable.name} = ${varInput}`;
    });

    $('#usageBuilderOutput').html(`${additionalContent}module "${moduleDetails.name}" {
  source  = "${window.location.hostname}/${analytics_token}__${moduleDetails.module_provider_id}"
  version = "${moduleDetails.terraform_example_version_string}"
${outputTf}
}`);
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

    // If current version is not available or there are no
    // versions, set warning and exit
    if (! moduleDetails.versions.length) {
        showNoAvailableVersions();
        return;
    }

    showModuleDetailsBody();
    enableTerraregExclusiveTags();
    setProviderLogo(moduleDetails);

    setModuleTitle(moduleDetails);
    setModuleProvider(moduleDetails);
    setModuleDescription(moduleDetails);
    setPublishedAt(moduleDetails);
    setOwner(moduleDetails);
    setSourceUrl(moduleDetails.display_source_url);

    addModuleLabels(moduleDetails, $("#module-title"));
}

class ReadmeTab extends BaseTab {
    constructor(readmeUrl) {
        super();
        this._readmeUrl = readmeUrl;
    }
    get name() {
        return 'readme';
    }
    render() {
        this._renderPromise = new Promise((resolve) => {
            if (!this._readmeUrl) {
                resolve(false);
                return;
            }
            $.get(this._readmeUrl, (readmeContent) => {
                // If no README is defined, exit early
                if (!readmeContent) {
                    resolve(false);
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
                $("#module-tab-link-readme").removeClass('default-hidden');
            
                resolve(true);
            });
        });
    }
}

class TabFactory {
    constructor() {
        this._tabs = [];
        this._tabsLookup = {};
    }
    registerTab(tab) {
        if (this._tabs[tab.name] !== undefined) {
            throw "Tab already registered";
        }
        this._tabsLookup[tab.name] = tab;
        this._tabs.push(tab);
    }
    renderTabs() {
        this._tabs.forEach((tab) => {
            tab.render();
        });
    }
    async setDefaultTab() {

        // Check if tab is defined in page URL anchor and if it's
        // a valid tab
        let windowHashValue = $(location).attr('hash').replace('#', '');
        if (windowHashValue && this._tabsLookup[windowHashValue] !== undefined) {
            let tab = this._tabsLookup[windowHashValue];

            // Check if tab has successfully loaded
            let isValid = await tab.isValid();

            if (isValid == true) {
                // Load tab and return
                selectModuleTab(tab.name, false);
                return;
            }
        }

        // Iterate through all tabs and select first valid tab
        for (const tab of this._tabs) {
            let isValid = await tab.isValid();
            if (isValid == true) {
                selectModuleTab(tab.name, false);
                return;
            }
        }
    }
}

/*
 * Setup elements of page used by root module pages.
 *
 * @param data Data from router
 */
async function setupRootModulePage(data) {
    let id = getCurrentObjectId(data);

    let moduleDetails = await getModuleDetails(id);

    let tabFactory = new TabFactory();

    if (moduleDetails.versions.length) {
        // Register tabs in order of being displayed to user, by default
        tabFactory.registerTab(new ReadmeTab(`/v1/terrareg/modules/${moduleDetails.id}/readme_html`));
        tabFactory.registerTab(new InputsTab(moduleDetails.root));
    }

    tabFactory.registerTab(new SettingsTab(moduleDetails));
    tabFactory.registerTab(new IntegrationsTab(moduleDetails));

    if (moduleDetails.versions.length) {

        tabFactory.registerTab(new OutputsTab(moduleDetails.root));
        tabFactory.registerTab(new ProvidersTab(moduleDetails.root));
        tabFactory.registerTab(new ResourcesTab(moduleDetails.root));
        tabFactory.registerTab(new AnalyticsTab(moduleDetails));
        tabFactory.registerTab(new UsageBuilderTab(moduleDetails));

        populateVersionSelect(moduleDetails);
        setupModuleVersionDeletionSetting(moduleDetails);
        populateSubmoduleSelect(moduleDetails);
        populateExampleSelect(moduleDetails);
        populateTerraformUsageExample(moduleDetails);
        populateDownloadSummary(moduleDetails);
    }

    tabFactory.renderTabs();
    tabFactory.setDefaultTab();
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

    populateCurrentSubmodule(`Submodule: ${submodulePath}`)
    populateVersionText(moduleDetails);
    populateTerraformUsageExample(moduleDetails, submodulePath);
    enableBackToParentLink(moduleDetails);

    let tabFactory = new TabFactory();

    tabFactory.registerTab(new ReadmeTab(`/v1/terrareg/modules/${moduleDetails.id}/submodules/readme_html/${submodulePath}`));
    tabFactory.registerTab(new InputsTab(submoduleDetails));
    tabFactory.registerTab(new OutputsTab(submoduleDetails));
    tabFactory.registerTab(new ProvidersTab(submoduleDetails));
    tabFactory.registerTab(new ResourcesTab(submoduleDetails));
    tabFactory.renderTabs();
    tabFactory.setDefaultTab();
}

/*
 * Setup elements of page for example
 *
 * @param data Data from router
 */
async function setupExamplePage(data) {
    let moduleVersionId = getCurrentObjectId(data);
    let moduleDetails = await getModuleDetails(moduleVersionId);

    let examplePath = data.undefined;
    let submoduleDetails = await getExampleDetails(moduleDetails.id, examplePath);

    populateCurrentSubmodule(`Example: ${examplePath}`)
    populateVersionText(moduleDetails);
    enableBackToParentLink(moduleDetails);

    let tabFactory = new TabFactory();

    tabFactory.registerTab(new ReadmeTab(`/v1/terrareg/modules/${moduleDetails.id}/examples/readme_html/${examplePath}`));
    tabFactory.registerTab(new ExampleFilesTab(moduleDetails, submoduleDetails));
    tabFactory.registerTab(new InputsTab(submoduleDetails));
    tabFactory.registerTab(new OutputsTab(submoduleDetails));
    tabFactory.registerTab(new ProvidersTab(submoduleDetails));
    tabFactory.registerTab(new ResourcesTab(submoduleDetails));
    tabFactory.renderTabs();
    tabFactory.setDefaultTab();
}


function createBreadcrumbs(data) {
    let breadcrumbs = [
        ["Modules", "modules"],
        [data.namespace, data.namespace],
        [data.module, data.module],
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
