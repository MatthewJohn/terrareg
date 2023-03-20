
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
    async renderTabs() {
        for (const tab of this._tabs) {
            tab.render();
        }
        for (const tab of this._tabs) {
            await tab._renderPromise;
        }
    }

    async setDefaultTab() {

        // Check if tab is defined in page URL anchor and if it's
        // a valid tab
        let windowHashValue = $(location).attr('hash').replace('#', '');
        if (! windowHashValue) {
            return;
        }
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

        // Check for any elements that have a child element that have ID/name of the anchor
        for (const tab of this._tabs) {
            let elements = $.find(`#module-tab-${tab.name} #${windowHashValue}, #module-tab-${tab.name} [name="${windowHashValue}"]`);
            if (elements.length) {
                let element = elements[0];

                // Select tab
                selectModuleTab(tab.name, false);

                // Scroll to element
                element.scrollIntoView();
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

class BaseTab {
    constructor() {
        this._renderPromise = undefined;
    }
    render() {  }
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
    render() { }
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


class AdditionalTab extends BaseTab {
    constructor(tabName, fileUrl) {
        super();
        this._tabName = tabName;
        this._fileUrl = fileUrl;
        this._name = 'custom-' + this._tabName.replace(/[^a-zA-Z]/i, '-');
    }
    get name() {
        return this._name;
    }
    render() {
        this._renderPromise = new Promise((resolve) => {
            if (!this._fileUrl) {
                resolve(false);
                return;
            }
            $.get(this._fileUrl, (fileContent) => {
                if (!fileContent) {
                    resolve(false);
                    return;
                }

                let tab = $(`<div id="module-tab-${this.name}" class="module-tabs content default-hidden"></div>`);

                // Populate file conrtent
                tab.append(fileContent);

                // Add 'table' class to all tables in README
                tab.find("table").addClass("table");

                // Replace size of headers
                tab.find("h1").addClass("subtitle").addClass("is-3");
                tab.find("h2").addClass("subtitle").addClass("is-4");
                tab.find("h3").addClass("subtitle").addClass("is-5");
                tab.find("h4").addClass("subtitle").addClass("is-6");

                // Insert tab link content
                let insertAfterContent = $('#module-tab-resources')[0];
                tab.insertAfter(insertAfterContent);

                // Insert tab link
                let tabLink = $(`<li id="module-tab-link-${this.name}" class="module-tab-link"><a onclick="selectModuleTab('${this.name}')">${this._tabName}</a></li>`);
                let insertAfterLink = $('#module-tab-link-resources');
                tabLink.insertAfter(insertAfterLink);

                resolve(true);
            });
        });
    }
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
                if (!data.length) {
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
                versionTd.append(provider.version);
                providerRow.append(versionTd);

                providerTabTbody.append(providerRow);
            });

            // Show tab link
            $('#module-tab-link-providers').removeClass('default-hidden');
            resolve(true);
        });
    }
}

class SecurityIssuesTab extends ModuleDetailsTab {
    get name() {
        return 'security-issues';
    }
    async render() {
        this._renderPromise = new Promise(async (resolve) => {
            if (this._moduleDetails.security_results) {
                let tfsecTab = $("#module-tab-security-issues");
                let tfsecTabTbody = tfsecTab.find("tbody");
                this._moduleDetails.security_results.forEach((tfsec) => {
                    let tfsecRow = $("<tr></tr>");

                    let blankTd = $('<td class="is-vcentered"></td>');
                    tfsecRow.append(blankTd);

                    let color = 'darkgray';
                    if (tfsec.severity == "CRITICAL") {
                        color = 'red';
                    } else if (tfsec.severity == "HIGH") {
                        color = 'orangered';
                    } else if (tfsec.severity == "MEDIUM") {
                        color = 'orange';
                    }
                    let severityTd = `<td class="is-vcentered"><span class="tag is-primary is-light" style="background-color: ${color}; color: white">${tfsec.severity}</span></td>`;
                    tfsecRow.append(severityTd);

                    let fileTd = $('<td class="is-vcentered"></td>');
                    if (tfsec.location && tfsec.location.filename) {
                        fileTd.text(tfsec.location.filename);
                    }
                    tfsecRow.append(fileTd);

                    let descriptionTd = $('<td class="is-vcentered"></td>');
                    descriptionTd.text(tfsec.rule_description);
                    tfsecRow.append(descriptionTd);

                    let ruleidTd = $('<td class="is-vcentered"></td>');
                    let tfsec_link = '#';
                    if (tfsec.links && tfsec.links[0]) {
                        tfsec_link = tfsec.links[0];
                    }
                    ruleidTd.html(`<a href="${tfsec_link}" target="_blank" rel="noopener noreferrer">${tfsec.rule_id}</a>`);
                    tfsecRow.append(ruleidTd);

                    let providerTd = $('<td class="is-vcentered"></td>');
                    providerTd.text(tfsec.rule_provider);
                    tfsecRow.append(providerTd);

                    let serviceTd = $('<td class="is-vcentered"></td>');
                    serviceTd.text(tfsec.rule_service);
                    tfsecRow.append(serviceTd);

                    let resourceTd = $('<td class="is-vcentered"></td>');
                    resourceTd.text(tfsec.resource);
                    tfsecRow.append(resourceTd);

                    let startLineTd = $('<td class="is-vcentered"></td>');
                    if (tfsec.location) {
                        startLineTd.text(tfsec.location.start_line);
                    }
                    tfsecRow.append(startLineTd);

                    let endLineTd = $('<td class="is-vcentered"></td>');
                    if (tfsec.location) {
                        endLineTd.text(tfsec.location.end_line);
                    }
                    tfsecRow.append(endLineTd);

                    let impactTd = $('<td class="is-vcentered"></td>');
                    impactTd.text(tfsec.impact);
                    tfsecRow.append(impactTd);

                    let resolutionTd = $('<td class="is-vcentered"></td>');
                    resolutionTd.text(tfsec.resolution);
                    tfsecRow.append(resolutionTd);

                    let resourcesTd = $('<td class="is-vcentered"></td>');
                    resourcesTd.html('<br/>');
                    if (tfsec.links) {
                        for (var i = 0; i < tfsec.links.length; i++)
                        {
                            resourcesTd.append(` - <a href="${tfsec.links[i]}" target="_blank" rel="noopener noreferrer">${tfsec.links[i]}</a><br/>`);
                        }
                    }
                    tfsecRow.append(resourcesTd);

                    tfsecTabTbody.append(tfsecRow);
                });

                // Show tab link
                $('#module-tab-link-security-issues').removeClass('default-hidden');
            }

            $("#security-issues-table").DataTable({
                order: [[1, 'asc']],
                autoWidth: false,
                columnDefs: [
                    {
                        targets: [2, 4, 5, 6, 7, 8, 9, 10, 11, 12],
                        className: "none"
                    },
                    {
                        targets: [1],
                        width: "1%"
                    },
                ],
                rowGroup: {
                    dataSrc: [2]
                },
                pagingType: "full_numbers_no_ellipses",
                lengthMenu: [
                    [25, 50, -1],
                    [25, 50, 'All'],
                ],
            });

            resolve(true);
        });
    }
}

class ResourcesTab extends ModuleDetailsTab {
    constructor(moduleDetails, graphUrl) {
        super(moduleDetails);
        this._graphUrl = graphUrl;
    }
    get name() {
        return 'resources';
    }
    async render() {
        this._renderPromise = new Promise(async (resolve) => {
            // Populate link to resources graph
            $('#resourceDependencyGraphLink').attr("href", this._graphUrl);

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

class IntegrationsTab extends ModuleDetailsTab {
    get name() {
        return 'integrations';
    }
    async render() {
        this._renderPromise = new Promise(async (resolve) => {
            let config = await getConfig();
            let loggedIn = await isLoggedIn();

            if (config.ALLOW_MODULE_HOSTING) {
                $("#module-integrations-upload-container").removeClass('default-hidden');
            }
            if (!config.PUBLISH_API_KEYS_ENABLED ||
                    loggedIn && (
                        loggedIn.site_admin ||
                        Object.keys(loggedIn.namespace_permissions).indexOf(this._moduleDetails.namespace) !== -1
                    )
            ) {
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
                    // Add URL to code block, generating from
                    // current protocol, hostname and URL endpoint in integration.
                    urlCodeContent += pathToUrl(integration.url);
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

class SettingsTab extends ModuleDetailsTab {
    get name() {
        return 'settings';
    }
    async render() {
        this._renderPromise = new Promise(async (resolve) => {

            let loggedIn = await isLoggedIn();
            // Return immediately if user is not logged in
            if (!loggedIn) {
                resolve(false);
                return;
            }

            if (this._moduleDetails.verified) {
                $('#settings-verified').attr('checked', true);
            }

            // Check if namespace is auto-verified and, if so, show message
            getNamespaceDetails(this._moduleDetails.namespace).then((namespaceDetails) => {
                if (namespaceDetails.is_auto_verified) {
                    $('#settings-verified-auto-verified-message').removeClass('default-hidden');
                }
            });

            // Setup git providers
            let config = await getConfig();
            let gitProviderSelect = $('#settings-git-provider');

            if (config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER) {
                let customGitProvider = $('<option></option>');
                customGitProvider.val('');
                customGitProvider.text('Custom');
                if (!this._moduleDetails.git_provider_id) {
                    customGitProvider.attr('selected', '');
                }
                gitProviderSelect.append(customGitProvider);

                // Enable custom git provider inputs
                $.find('.settings-custom-git-provider-container').forEach((inputContainer) => {
                    $(inputContainer).removeClass('default-hidden');
                });
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

            $('#settings-git-path').val(this._moduleDetails.git_path);
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

            // Show settings tab
            $('#module-tab-link-settings').removeClass('default-hidden');
            resolve(true);
        });
    }
}

class UsageBuilderRowFactory {
    constructor(terraregConfig) {
        this.terraregConfig = terraregConfig;
    }
    getRowFromConfig(config) {
        switch (config.type) {
            case "text": {
                return new UsageBuilderTextRow(config);
            }
            case "list": {
                return new UsageBuilderListRow(config);
            }
            case "number": {
                return new UsageBuilderNumberRow(config);
            }
            case "boolean": {
                return new UsageBuilderBooleanRow(config);
            }
            case "select": {
                return new UsageBuilderSelectRow(config);
            }
            case "static": {
                return new UsageBuilderStaticRow(config);
            }
            default: {
                console.log('Unknown usage builder row type:', config.type);
                break;
            }
        }
    }
    getAnalyticsRow() {
        return new UsageBuilderAnalyticstokenRow(this.terraregConfig);
    }
}

class BaseUsageBuilderRow {
    constructor(config) {
        this.config = config;
        this._inputRow = undefined;
    }

    get name() {
        return this.config.name;
    }

    get inputId() {
        return `usageBuilderInput-${this.name}`;
    }
    get inputIdHash() {
        return `#${this.inputId}`;
    }

    get required() {
        return this.config.required;
    }

    getInputRow() {

        let inputRow = $('<tr></tr>');

        let inputNameTd = $('<td></td>');
        inputNameTd.attr('style', 'width: 10%');
        inputNameTd.text(this.name);
        inputRow.append(inputNameTd);

        let requiredTd = $('<td></td>');
        inputNameTd.attr('style', 'width: 10%');
        requiredTd.text(this.required === true ? "Yes" : "No");
        inputRow.append(requiredTd);

        let additionalHelpTd = $('<td></td>');
        additionalHelpTd.attr('style', 'width: 50%');
        additionalHelpTd.text(this.config.additional_help ? this.config.additional_help : '');
        inputRow.append(additionalHelpTd);

        let valueTd = $('<td></td>');
        valueTd.attr('style', 'width: 20%');

        this.generateInputDiv(valueTd);
        inputRow.append(valueTd);
        this._inputRow = inputRow;

        return inputRow;
    }

    quoteString(input) {
        // Place input value directly into double quotes.
        // Escape backslashes and then escape double quotes.
        if (this.config.quote_value && typeof input !== "undefined") {
            return '"' + input.replace(/\\/g, '\\\\').replace(/"/g, '\\"') + '"';
        } else {
            return input;
        }
    }

    getTerraformContent(optionalEnabled) {
        if (this.config.required === false && optionalEnabled === false) {
            return {
                'body': '',
                'additionalContent': ''
            };
        }

        return this._getTerraformContent()
    }
}

class UsageBuilderStaticRow extends BaseUsageBuilderRow {
    getInputRow() {
        return null;
    }

    _getTerraformContent() {
        let varInput = this.quoteString(this.config.value);
        return {
            'body': `\n  ${this.name} = ${varInput}`,
            'additionalContent': ''
        };
    }
}

class UsageBuilderTextRow extends BaseUsageBuilderRow {
    generateInputDiv(valueTd) {
        let placeholder = this.config.default_value
        if (typeof placeholder === 'object') {
            placeholder = JSON.stringify(placeholder)
        }
        let inputDiv = $('<input />');
        inputDiv.addClass('input');

        inputDiv.attr('type', 'text');

        inputDiv.attr('id', this.inputId);
        inputDiv.attr('placeholder', placeholder);
        valueTd.append(inputDiv);
    }

    _getTerraformContent() {
        let userValue = this._inputRow.find(this.inputIdHash).val()
        let varInput = this.quoteString(userValue);

        if (this.config.required === false && userValue === "") {
            return {
                'body': '',
                'additionalContent': ''
            };
        }
        return {
            'body': `\n  ${this.name} = ${varInput}`,
            'additionalContent': ''
        };
    }
}

class UsageBuilderNumberRow extends BaseUsageBuilderRow {
    generateInputDiv(valueTd) {
        let placeholder = this.config.default_value
        if (typeof placeholder === 'object') {
            placeholder = JSON.stringify(placeholder)
        }
        let inputDiv = $('<input />');
        inputDiv.addClass('input');

        inputDiv.attr('type', 'number');

        inputDiv.attr('id', this.inputId);
        inputDiv.attr('placeholder', placeholder);
        valueTd.append(inputDiv);
    }

    _getTerraformContent() {
        let varInput = this._inputRow.find(this.inputIdHash).val();

        if (this.config.required === false && varInput == '') {
            return {
                'body': '',
                'additionalContent': ''
            };
        }
        return {
            'body': `\n  ${this.name} = ${varInput}`,
            'additionalContent': ''
        };
    }
}

class UsageBuilderListRow extends BaseUsageBuilderRow {
    generateInputDiv(valueTd) {
        let placeholder = this.config.default_value
        if (typeof placeholder === 'object') {
            placeholder = JSON.stringify(placeholder)
        }
        let inputDiv = $('<input />');
        inputDiv.addClass('input');
        inputDiv.attr('type', 'text');

        // Add class for input ID
        inputDiv.addClass(this.inputId);

        // Append 0 to input ID, for first input for list
        let inputId = this.inputId + '0';
        inputDiv.on('keyup', () => {

            // Check all list inputs, remove any empty ones
            // and add an additional input, if last input is populated
            let inputIdName = `usageBuilderInput-${this.name}`;
            let listInputDivs = $(`.${inputIdName}`);
            for (const inputDiv of listInputDivs) {
                let val = inputDiv.value;

                // Check if input div is last item
                if (listInputDivs.index(inputDiv) == (listInputDivs.length - 1)) {

                    // If input contains a value, clone and create new input
                    if (val) {
                        let newInput = $(inputDiv).clone();

                        // Update Id of new input
                        newInput.attr('id', inputIdName + listInputDivs.length);

                        // Reset value of new iput
                        newInput.val('');

                        // Bind original keyup method to new input div
                        newInput.on('keyup', $._data(inputDiv).events.keyup[0].handler);

                        // Add new input after the original
                        newInput.insertAfter(inputDiv);
                    }
                } else {
                    // Otherwise, check if item is empty
                    if (!val) {
                        inputDiv.remove();
                    }
                }
            }
        });

        inputDiv.attr('id', inputId);
        inputDiv.attr('placeholder', placeholder);
        valueTd.append(inputDiv);
    }

    _getTerraformContent() {
        let valueList = [];

        let listInputDivs = this._inputRow.find(`.${this.inputId}`);
        for (const inputDiv of listInputDivs) {
            let val = inputDiv.value;

            // Check if input div is last item
            if (listInputDivs.index(inputDiv) == (listInputDivs.length - 1)) {

                // If input contains a value add to valueList
                if (val) {

                    // Add value of current input to list
                    valueList.push(this.quoteString(val));
                }
            } else {
                // If input contains a value add to valueList
                if (val) {
                    valueList.push(this.quoteString(val));
                }
            }
        }

        let varInput = `[${valueList.join(', ')}]`;
        if (this.config.required === false && valueList.length == 0) {
            return {
                'body': '',
                'additionalContent': ''
            };
        }
        return {
            'body': `\n  ${this.name} = ${varInput}`,
            'additionalContent': ''
        };
    }
}

class UsageBuilderBooleanRow extends BaseUsageBuilderRow {
    constructor(config) {
        super(config);
        this._inputShown = true;
    }
    generateInputDiv(valueTd, forceShowCheckbox=false) {
        // Remove all child elements from the inputDiv
        valueTd.empty();

        // If the config is not required and default value is null,
        // show code default value with a link to enable modification,
        // which will re-call this function, forcing the display of the checkbox input
        if (this.config.required === false && this.config.default_value == null && forceShowCheckbox === false) {
            this._inputShown = false;

            let nullText = $('<code>null</code>');
            valueTd.append(nullText);
            valueTd.append('<br />');

            let changeLink = $('<a>Modify</a>');
            changeLink.bind('click', () => {
                this.generateInputDiv(valueTd, true);
            });
            valueTd.append(changeLink);
            return;
        }
        this._inputShown = true;

        let inputDiv = $('<input />');
        inputDiv.addClass('checkbox');
        inputDiv.attr('type', 'checkbox');
        inputDiv.attr('id', this.inputId);

        if (this.config.default_value == true) {
            inputDiv.prop( "checked", true )
        }
        valueTd.append(inputDiv);

        // Add link to remove input checkbox for optional value
        if (this.config.required == false && this.config.default_value == null) {
            valueTd.append('<br />');
            let changeLink = $('<a>Undo</a>');
            changeLink.bind('click', () => {
                this.generateInputDiv(valueTd);
            });
            valueTd.append(changeLink);
        }
    }

    _getTerraformContent() {
        // If input is now shown, skip the row
        if (!this._inputShown) {
            return {
                'body': '',
                'additionalContent': ''
            };
        }

        let is_checked = this._inputRow.find(this.inputIdHash).prop('checked');

        // If default value is shown (i.e. the value is optional),
        // do not show any Terraform output
        if (is_checked === this.config.default_value) {
            return {
                'body': '',
                'additionalContent': ''
            };
        }
        let varInput = JSON.stringify(is_checked);

        return {
            'body': `\n  ${this.name} = ${varInput}`,
            'additionalContent': ''
        };
    }
}

class UsageBuilderSelectRow extends BaseUsageBuilderRow {
    generateInputDiv(valueTd) {
        let inputDiv = $('<div></div>');
        inputDiv.addClass('select');
        let inputSelect = $('<select></select>');
        let inputId = this.inputId;
        inputSelect.attr('id', inputId);
        this.config.choices.forEach((inputChoice, itx) => {
            // If choices is list of strings, use the string as the name,
            // otherwise, use name attribute of object.
            let inputName = typeof inputChoice === 'string' ? inputChoice : inputChoice.name;
            let option = $('<option></option>');
            option.val(itx);
            option.text(inputName);
            inputSelect.append(option);
        });
        // If custom input is available, add to select
        if (this.config.allow_custom) {
            inputSelect.on('change', function () {
                var selectedText  = this.selectedOptions[0].value;
                inputId = $(this).prop('id');
                let customInputId = `#${inputId}-customValue`;
                // Check if custom type
                if (selectedText == 'custom') {
                    // Display custom text input
                    $(customInputId).attr("style", "display:block");
                } else {
                    // Hide custom input and clear value
                    $(customInputId).attr("style", "display:none")
                    $(customInputId).value = '';
                }
            });
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
        customInput.css('display', 'none');
        valueTd.append(customInput);
    }

    _getTerraformContent() {
        let userInput = '';
        let varInput = '';
        let additionalContent = '';


        let selectIndex = this._inputRow.find(this.inputIdHash).val();
        let customInputId = `${this.inputId}-customValue`;

        // Check if custom type
        if (selectIndex == 'custom') {

            // Use value of custom input as output
            userInput = this._inputRow.find(customInputId).val()
        } else {
            // If choice is a string, add the choice name to as the value
            if (typeof this.config.choices[selectIndex] === 'string') {
                userInput = this.config.choices[selectIndex];
            } else {
                // Otherwise, use the attribute for the value
                userInput = this.config.choices[selectIndex].value;

                // If object has additional_content, add it to the TF output
                if (this.config.choices[selectIndex].additional_content) {
                    additionalContent += this.config.choices[selectIndex].additional_content + '\n\n';
                }
            }
        }
        varInput = this.quoteString(userInput)

        if (this.config.required === false && userInput == "") {
            return {
                'body': '',
                'additionalContent': ''
            };
        }
        return {
            'body': `\n  ${this.name} = ${varInput}`,
            'additionalContent': additionalContent
        };
    }
}

class UsageBuilderAnalyticstokenRow extends BaseUsageBuilderRow {
    constructor(terraregConfig) {
        super(null);
        this.terraregConfig = terraregConfig;
        this._inputDiv = undefined;
    }
    getInputRow() {
        // Setup analytics input row
        let analyticsTokenInputRow = $('<tr></tr>');

        let analyticsTokenName = $('<td></td>');
        analyticsTokenName.text(this.terraregConfig.ANALYTICS_TOKEN_PHRASE);
        analyticsTokenInputRow.append(analyticsTokenName);

        let analyticsrequiredTd = $('<td></td>');
        analyticsrequiredTd.text("Yes");
        analyticsTokenInputRow.append(analyticsrequiredTd);

        let analyticsTokenDescription = $('<td></td>');
        analyticsTokenDescription.text(this.terraregConfig.ANALYTICS_TOKEN_DESCRIPTION);
        analyticsTokenInputRow.append(analyticsTokenDescription);

        let analyticsTokenInputTd = $('<td></td>');
        let analyticsTokenInputField = $('<input />');
        analyticsTokenInputField.attr('class', 'input');
        analyticsTokenInputField.attr('id', 'usageBuilderAnalyticsToken');
        analyticsTokenInputField.attr('type', 'text');
        analyticsTokenInputField.attr('placeholder', this.terraregConfig.EXAMPLE_ANALYTICS_TOKEN);
        this._inputDiv = analyticsTokenInputField;
        analyticsTokenInputTd.append(analyticsTokenInputField);
        analyticsTokenInputRow.append(analyticsTokenInputTd);

        return analyticsTokenInputRow;
    }

    getValue() {
        if (this._inputDiv) {
            return this._inputDiv.val();
        }
        return "";
    }
}


class UsageBuilderTab extends ModuleDetailsTab {
    constructor(moduleDetails) {
        super(moduleDetails);
        this._inputRows = [];
        this._analyticsInput = undefined;
    }
    get name() {
        return 'usage-builder';
    }
    async render() {
        this._renderPromise = new Promise(async (resolve) => {
            let usageBuilderTable = $('#usageBuilderTable');

            let inputVariables = await getUsageBuilderVariables(this._moduleDetails.id);

            if (inputVariables === null || !inputVariables.length) {
                // If there are no variables present in the usage builder,
                // then exit early
                resolve(false);
                return;
            }

            let config = await getConfig();

            let usageBuilderRowFactory = new UsageBuilderRowFactory(config);

            this._analyticsInput = usageBuilderRowFactory.getAnalyticsRow();
            usageBuilderTable.append(this._analyticsInput.getInputRow());

            // Build input table
            inputVariables.forEach((inputVariable) => {

                let inputRowObject = usageBuilderRowFactory.getRowFromConfig(inputVariable);
                if (inputRowObject) {
                    this._inputRows.push(inputRowObject);
                    let inputRow = inputRowObject.getInputRow();
                    if (inputRow) {
                        usageBuilderTable.append(inputRow);
                    }
                }
            });
            globalThis.usageBuilderUseOptional = false
            globalThis.moduleDetails = this._moduleDetails
            globalThis.usageBuilderTab = this;
            globalThis.usageBuilderTable = $("#usage-builder-table").DataTable({
                columnDefs: [
                    {
                        targets: [0, 1],
                        width: "1%"
                    },
                ],
                order: [[1, 'desc'], [0, 'asc']],
                ordering: false,
                // pagingType: "full_numbers_no_ellipses",
                lengthMenu: [
                    [10, 25, 50, -1],
                    [10, 25, 50, 'All'],
                ],
                // dom: 'Bfrtip',
                buttons: {
                    dom: {
                        button: {
                            tag: 'button',
                            className: 'button is-outlined'
                        }
                    },
                    buttons: [
                        {
                            text: 'Show Optional Variables',
                            className: 'is-dark',
                            action: function (e, dt, node, config) {
                                if (dt.column(1).search() === 'Yes') {
                                    this.text('Hide Optional Variables');
                                    dt.column(1).search('').draw(true);
                                    globalThis.usageBuilderUseOptional = true
                                } else {
                                    this.text('Show Optional Variables');
                                    dt.column(1).search('Yes').draw(true);
                                    globalThis.usageBuilderUseOptional = false
                                }
                            }
                        },
                        {
                            text: 'Generate Terraform',
                            className: 'is-info',
                            action: function ( e, dt, node, conf ) {
                                e.preventDefault();
                                globalThis.usageBuilderTab.updateUsageBuilderOutput()
                            }
                        }
                    ],
                },
                searchCols: [
                    null,
                    { search: "Yes" },
                ]
            });

            globalThis.usageBuilderTable.buttons().container()
                .appendTo( $('div.column.is-full', globalThis.usageBuilderTable.table().container()).eq(0) );

            // Show tab
            $('#module-tab-link-usage-builder').removeClass('default-hidden');
            $('#example-link-usage-builder').removeClass('default-hidden');
            resolve(true);

        });
    }

    async updateUsageBuilderOutput() {
        let outputTf = '';
        let additionalContent = '';
        let moduleDetails = globalThis.moduleDetails;
    
        let analytics_token = this._analyticsInput.getValue();
    
        for (let inputRow of this._inputRows) {
            let content = inputRow.getTerraformContent(globalThis.usageBuilderUseOptional);
            outputTf += content.body
            additionalContent += content.additionalContent;
        }
        $('#usageBuilderOutput').html(`${additionalContent}module "${moduleDetails.name}" {
  source  = "${window.location.hostname}/${analytics_token}__${moduleDetails.module_provider_id}"
  version = "${moduleDetails.terraform_example_version_string}"
${outputTf}
}`);
    }
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
    let currentIsLatestVersion = false;

    let userPreferences = getUserPreferences();

    $.get(`/v1/terrareg/modules/${moduleDetails.module_provider_id}/versions` +
        `?include-beta=${userPreferences["show-beta-versions"]}&` +
        `include-unpublished=${userPreferences["show-unpublished-versions"]}`).then((versions) => {
            let foundLatest = false;
            for (let versionDetails of versions) {
                let versionOption = $("<option></option>");
                let isLatest = false;

                // Set value of option to view URL of module version
                versionOption.val(`/modules/${moduleDetails.namespace}/${moduleDetails.name}/${moduleDetails.provider}/${versionDetails.version}`);

                let versionText = versionDetails.version;
                // Add '(latest)' suffix to the first (latest) version
                if (foundLatest == false && versionDetails.beta == false && versionDetails.published == true) {
                    versionText += " (latest)";
                    foundLatest = true;
                    isLatest = true;
                } else {
                    if (versionDetails.beta) {
                        versionText += ' (beta)';
                    }
                    if (versionDetails.published == false) {
                        versionText += ' (unpublished)';
                    }
                }
                versionOption.text(versionText);

                // Set version that matches current module to selected
                if (versionDetails.version == moduleDetails.version) {
                    versionOption.attr("selected", "");
                    currentVersionFound = true;

                    // Determine if the current version is the latest version
                    // (first in list of versions)
                    if (isLatest) {
                        currentIsLatestVersion = true;
                    }
                }

                versionSelection.append(versionOption);
            }

            // If current version has not been found, add fake version to drop-down
            if (currentVersionFound == false) {
                let versionOption = $("<option></option>");
                let suffix = '';
                if (moduleDetails.beta) {
                    suffix += ' (beta)';
                }
                if (!moduleDetails.published) {
                    suffix += ' (unpublished)';
                }
                versionOption.text(`${moduleDetails.version}${suffix}`);
                versionOption.attr("selected", "");
                versionSelection.append(versionOption);

            }
            if (!currentIsLatestVersion && !moduleDetails.beta && moduleDetails.published) {
                // Otherwise, if user is not viewing the latest version,
                // display warning
                $("#non-latest-version-warning").removeClass('default-hidden');
            }

            if (moduleDetails.beta) {
                // If the beta module is published, show warning about
                // beta version and how to use it.
                if (moduleDetails.published) {
                    $("#beta-warning").removeClass('default-hidden');
                }
            }
            if (!moduleDetails.published) {
                // Add warning to page about unpublished version
                $("#unpublished-warning").removeClass('default-hidden');
            }

            // Show version drop-down
            $('#details-version').removeClass('default-hidden');
        });
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
 * Set custom links
 *
 * @param moduleDetails
 */
function populateCustomLinks(moduleDetails) {
    let customLinkParent = $('#custom-links');
    for (let linkDetails of moduleDetails.custom_links) {
        let link = $('<a></a>');
        link.addClass('custom-link');
        link.attr('href', linkDetails.url);
        link.text(linkDetails.text);
        customLinkParent.append(link);
        customLinkParent.append('<br />');
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
async function populateTerraformUsageExample(moduleDetails, submoduleDetails) {
    let config = await getConfig();

    // Check if module has been published - if not, do not
    // show usage example
    if (! moduleDetails.published) {
        return;
    }

    // Populate supported Terraform versions
    if (submoduleDetails.terraform_version_constraint) {
        $('#supported-terraform-versions-data').text(submoduleDetails.terraform_version_constraint);
        $('#supported-terraform-versions').removeClass('default-hidden');
    }

    // Update labels for example analytics token and phrase
    $("#usage-example-analytics-token").text(config.EXAMPLE_ANALYTICS_TOKEN);
    $("#usage-example-analytics-token-phrase").text(config.ANALYTICS_TOKEN_PHRASE);

    // Add example Terraform call to source section
    $("#usage-example-terraform").text(submoduleDetails.usage_example);

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

/*
 * Index new module provider version
 */
function indexModuleVersion(moduleDetails) {
    let moduleVersionToIndex = $("#indexModuleVersion").val();

    let inProgressMessage = $('#index-version-in-progress');
    let successMessage = $('#index-version-success');
    let errorMessage = $('#index-version-error');
    // Hide success/error and show in-progress
    successMessage.addClass('default-hidden');
    errorMessage.addClass('default-hidden');
    inProgressMessage.html('Indexing module version in progress...');
    inProgressMessage.removeClass('default-hidden');

    $.post(`/v1/terrareg/modules/${moduleDetails.module_provider_id}/${moduleVersionToIndex}/import`)
        .done(() => {
            // Show success message for importing module
            successMessage.html("Successfully indexed version");
            successMessage.removeClass('default-hidden');
            errorMessage.addClass('default-hidden');

            // If publish checkbox is checked, perform request to publish
            if ($("#indexModuleVersionPublish").is(":checked")) {
                inProgressMessage.html('Publishing module version in progress...');
                $.post(`/v1/terrareg/modules/${moduleDetails.module_provider_id}/${moduleVersionToIndex}/publish`)
                    .done(() => {
                        // If successful, update success message
                        successMessage.html("Successfully indexed and published version.");

                        // Hide in-progress
                        inProgressMessage.addClass('default-hidden');
                    })
                    .fail((res) => {
                        // Hide in-progress
                        inProgressMessage.addClass('default-hidden');

                        // Display any errors
                        errorMessage.html(failedResponseToErrorString(res));
                        errorMessage.removeClass('default-hidden');
                    });
            } else {
                // If publishing is not enabled, hide in-progress message
                inProgressMessage.addClass('default-hidden');
            }
        })
        .fail((res) => {
            // Render and show error
            errorMessage.html(failedResponseToErrorString(res));
            errorMessage.removeClass('default-hidden');
            // Hide in-progress
            inProgressMessage.addClass('default-hidden');
        });
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
    if (!config.DISABLE_TERRAREG_EXCLUSIVE_LABELS) {
        $.find('.terrareg-exclusive').forEach((tag) => {
            $(tag).removeClass('default-hidden');
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

    // Show module version deletion settings
    $('#module-version-delete-container').removeClass('default-hidden');
}

function deleteModuleVersion(moduleDetails) {
    $('#confirm-delete-module-version-div').removeClass('default-hidden');
    if ($('#confirm-delete-module-version').val() !== moduleDetails.version) {
        return;
    }
    $.ajax({
        url: `/v1/terrareg/modules/${moduleDetails.id}/delete`,
        method: 'delete',
        data: JSON.stringify({
            csrf_token: $('#settings-csrf-token').val()
        }),
        contentType: 'application/json'
    }).done(() => {
        // Reidrect to module provider page.
        // This will automatically redirect back
        // down the path to valid existing object.
        window.location.href = `/modules/${moduleDetails.module_provider_id}`;
    }).fail((res) => {
        if (res.status == 401) {
            $('#settings-status-error').html('You must be logged in to perform this action.<br />If you were previously logged in, please re-authentication and try again.');
        } else if (res.responseJSON && res.responseJSON.message) {
            $('#settings-status-error').html(res.responseJSON.message);
        } else {
            $('#settings-status-error').html('An unexpected error occurred');
        }
        $('#settings-status-error').removeClass('default-hidden');
        $(window).scrollTop($('#settings-status-error').offset().top);
    });
    return false;
}

function deleteModuleProvider(moduleDetails) {
    $('#confirm-delete-module-provider-div').removeClass('default-hidden');
    if ($('#confirm-delete-module-provider').val() !== moduleDetails.module_provider_id) {
        return;
    }
    $.ajax({
        url: `/v1/terrareg/modules/${moduleDetails.module_provider_id}/delete`,
        method: 'delete',
        data: JSON.stringify({
            csrf_token: $('#settings-csrf-token').val()
        }),
        contentType: 'application/json'
    }).done(() => {
        // Reidrect to module page.
        // This will automatically redirect back
        // down the path to valid existing object.
        window.location.href = `/modules/${moduleDetails.id}`;
    }).fail((res) => {
        if (res.status == 401) {
            $('#settings-status-error').html('You must be logged in to perform this action.<br />If you were previously logged in, please re-authentication and try again.');
        } else if (res.responseJSON && res.responseJSON.message) {
            $('#settings-status-error').html(res.responseJSON.message);
        } else {
            $('#settings-status-error').html('An unexpected error occurred');
        }
        $('#settings-status-error').removeClass('default-hidden');
        $(window).scrollTop($('#settings-status-error').offset().top);
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
            git_path: $('#settings-git-path').val(),
            verified: $('#settings-verified').is(':checked'),
            csrf_token: $('#settings-csrf-token').val()
        }),
        contentType: 'application/json'
    }).done(() => {
        $('#settings-status-success').removeClass('default-hidden');
    }).fail((res) => {
        if (res.status == 401) {
            $('#settings-status-error').html('You must be logged in to perform this action.<br />If you were previously logged in, please re-authentication and try again.');
        } else if (res.responseJSON && res.responseJSON.message) {
            $('#settings-status-error').html(res.responseJSON.message);
        } else {
            $('#settings-status-error').html('An unexpected error occurred');
        }
        $('#settings-status-error').removeClass('default-hidden');
        $(window).scrollTop($('#settings-status-error').offset().top);
    });

    // Return false to present default action
    return false;
}

function showSecurityWarnings(moduleDetails) {
    let securityIssuesContainer = $('#security-issues')
    if (moduleDetails.security_failures) {
        securityIssuesContainer.removeClass('default-hidden');
        $('#security-issues-text').text(`${moduleDetails.security_failures} Security issues`);
    } else {
        securityIssuesContainer.addClass('default-hidden');
    }
}

function showCostAnalysis(moduleDetails) {
    if (moduleDetails.cost_analysis && moduleDetails.cost_analysis.yearly_cost != null) {
        let yearlyCostContainer = $('#yearly-cost');
        yearlyCostContainer.removeClass('default-hidden');
        $('#yearly-cost-text').text(moduleDetails.cost_analysis.yearly_cost);

        // Setup label at top of module
        let yearCostLabel = $('#yearly-cost-label');
        yearCostLabel.find('#label-text').text(`$${moduleDetails.cost_analysis.yearly_cost}/yr`);
        yearCostLabel.removeClass('default-hidden');
    }
}

function showOutdatedExtractionDataWarning(moduleDetails) {
    if (moduleDetails.module_extraction_up_to_date === false) {
        $('#outdated-extraction-warning').removeClass('default-hidden');
    }
}

/*
 * Set HTML page title
 *
 * @param id Module id
 */
function setPageTitle(id) {
    document.title = `${id} - Terrareg`;
}


/*
 * Setup common elements of the page, shared between all types
 * of pages
 *
 * @param data Data from router
 */
async function setupBasePage(data) {

    let id = getCurrentObjectId(data);

    let moduleDetails = await getModuleDetails(id);

    // If current version is not available or there are no
    // versions, set warning and exit
    if (!moduleDetails.version) {
        showNoAvailableVersions();
        return;
    }

    showModuleDetailsBody();
    enableTerraregExclusiveTags();
    setProviderLogo(moduleDetails);

    setModuleTitle(moduleDetails);
    setModuleProvider(moduleDetails);

    addModuleLabels(moduleDetails, $("#module-labels"));
}

/*
 * Setup elements of page used by root module pages.
 *
 * @param data Data from router
 */
async function setupRootModulePage(data) {
    let id = getCurrentObjectId(data);

    let moduleDetails = await getModuleDetails(id);

    createBreadcrumbs(data);

    setPageTitle(moduleDetails.module_provider_id);
    setModuleDescription(moduleDetails);
    setPublishedAt(moduleDetails);
    setOwner(moduleDetails);

    let tabFactory = new TabFactory();

    if (moduleDetails.version) {
        // Register tabs in order of being displayed to user, by default
        tabFactory.registerTab(new ReadmeTab(`/v1/terrareg/modules/${moduleDetails.id}/readme_html`));
        tabFactory.registerTab(new InputsTab(moduleDetails.root));

        // Setup additional pages
        for (let tab_name of Object.keys(moduleDetails.additional_tab_files)) {
            tabFactory.registerTab(new AdditionalTab(tab_name, `/v1/terrareg/modules/${moduleDetails.id}/files/${moduleDetails.additional_tab_files[tab_name]}`));
        }
    }

    tabFactory.registerTab(new IntegrationsTab(moduleDetails));
    tabFactory.registerTab(new SettingsTab(moduleDetails));

    if (moduleDetails.version) {

        tabFactory.registerTab(new OutputsTab(moduleDetails.root));
        tabFactory.registerTab(new ProvidersTab(moduleDetails.root));
        tabFactory.registerTab(new ResourcesTab(moduleDetails.root, moduleDetails.graph_url));
        tabFactory.registerTab(new SecurityIssuesTab(moduleDetails));
        tabFactory.registerTab(new AnalyticsTab(moduleDetails));
        tabFactory.registerTab(new UsageBuilderTab(moduleDetails));

        showSecurityWarnings(moduleDetails);
        showOutdatedExtractionDataWarning(moduleDetails);
        populateVersionSelect(moduleDetails);
        setupModuleVersionDeletionSetting(moduleDetails);
        populateSubmoduleSelect(moduleDetails);
        populateExampleSelect(moduleDetails);
        populateTerraformUsageExample(moduleDetails, moduleDetails);
        populateDownloadSummary(moduleDetails);
        setSourceUrl(moduleDetails.display_source_url);
        populateCustomLinks(moduleDetails);
    }

    await tabFactory.renderTabs();
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

    createBreadcrumbs(data, submodulePath);

    setPageTitle(`${moduleDetails.module_provider_id}/${submoduleDetails.path}`);

    setModuleDescription(moduleDetails);
    setPublishedAt(moduleDetails);
    setOwner(moduleDetails);
    populateCurrentSubmodule(`Submodule: ${submodulePath}`)
    populateVersionText(moduleDetails);
    populateTerraformUsageExample(moduleDetails, submoduleDetails);
    enableBackToParentLink(moduleDetails);
    showSecurityWarnings(submoduleDetails);
    setSourceUrl(submoduleDetails.display_source_url);

    let tabFactory = new TabFactory();

    tabFactory.registerTab(new ReadmeTab(`/v1/terrareg/modules/${moduleDetails.id}/submodules/readme_html/${submodulePath}`));
    tabFactory.registerTab(new InputsTab(submoduleDetails));
    tabFactory.registerTab(new OutputsTab(submoduleDetails));
    tabFactory.registerTab(new ProvidersTab(submoduleDetails));
    tabFactory.registerTab(new ResourcesTab(submoduleDetails, submoduleDetails.graph_url));
    tabFactory.registerTab(new SecurityIssuesTab(submoduleDetails));
    await tabFactory.renderTabs();
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

    createBreadcrumbs(data, examplePath);

    setPageTitle(`${moduleDetails.module_provider_id}/${submoduleDetails.path}`);

    populateCurrentSubmodule(`Example: ${examplePath}`)
    populateVersionText(moduleDetails);
    enableBackToParentLink(moduleDetails);
    showSecurityWarnings(submoduleDetails);
    showCostAnalysis(submoduleDetails);
    setSourceUrl(submoduleDetails.display_source_url);

    let tabFactory = new TabFactory();

    tabFactory.registerTab(new ReadmeTab(`/v1/terrareg/modules/${moduleDetails.id}/examples/readme_html/${examplePath}`));
    tabFactory.registerTab(new ExampleFilesTab(moduleDetails, submoduleDetails));
    tabFactory.registerTab(new InputsTab(submoduleDetails));
    tabFactory.registerTab(new OutputsTab(submoduleDetails));
    tabFactory.registerTab(new ProvidersTab(submoduleDetails));
    tabFactory.registerTab(new ResourcesTab(submoduleDetails, submoduleDetails.graph_url));
    tabFactory.registerTab(new SecurityIssuesTab(submoduleDetails));
    await tabFactory.renderTabs();
    tabFactory.setDefaultTab();
}


async function createBreadcrumbs(data, subpath = undefined) {

    let namespaceName = data.namespace;
    let namespaceDetails = await getNamespaceDetails(namespaceName);
    if (namespaceDetails.display_name) {
        namespaceName = namespaceDetails.display_name;
    }

    let breadcrumbs = [
        ["Modules", "modules"],
        [namespaceName, data.namespace],
        [data.module, data.module],
        [data.provider, data.provider]
    ];
    if (data.version) {
        breadcrumbs.push([data.version, data.version]);
    }

    if (subpath !== undefined) {
        breadcrumbs.push([subpath, subpath]);
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
