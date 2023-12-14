
function timeDifference(previous) {
    var min = 60 * 1000;
    var hour = min * 60;
    var day = hour * 24;
    var month = day * 30;
    var year = day * 365;

    var current = new Date();
    var elapsed = current - previous;

    if (elapsed > year)
    {
        let cnt = Math.round(elapsed / year);
        if (cnt == 1)
        {
            return 'approximately ' + cnt + ' year ago';
        }
        return 'approximately ' + cnt + ' years ago';
    }
    else if (elapsed > month)
    {
        let cnt = Math.round(elapsed / month);
        if (cnt == 1)
        {
            return 'approximately ' + cnt + ' month ago';
        }
        return 'approximately ' + cnt + ' months ago';
    }
    else if (elapsed > day)
    {
        let cnt = Math.round(elapsed / day);
        if (cnt == 1)
        {
            return 'approximately ' + cnt + ' day ago'
        }
        return 'approximately ' + cnt + ' days ago';
    }
    else if (elapsed > hour)
    {
        let cnt = Math.round(elapsed / hour);
        if (cnt == 1)
        {
            return cnt + ' hour ago';
        }
        return cnt + ' hours ago';
    }
    else if (elapsed > min)
    {
        let cnt = Math.round(elapsed / min);
        if (cnt == 1)
        {
            return cnt + ' minute ago';
        }
        return cnt + ' minutes ago';
    }
    else
    {
        let cnt = Math.round(elapsed / 1000);
        if (cnt == 1)
        {
            return cnt + ' second ago';
        }
        return cnt + ' seconds ago';
    }
}

async function addProviderLogoTos(provider) {
    let provider_logos = await getProviderLogos();

    // Add provider TOS to results, if not already there
    if ($('#provider-tos-' + provider).length == 0) {
        let tos_object = document.createElement('p');
        tos_object.id = `provider-tos-${provider}`;
        tos_object.innerHTML = provider_logos[provider].tos;
        $('#provider-tos')[0].append(tos_object);
    }
}

class TerraformCompatibilityResult {
    constructor(text, color, icon) {
        this.text = text;
        this.color = color;
        this.icon = icon;
    }
}

function getTerraformCompatibilityResultObject(compatibilityResult) {
    if (compatibilityResult == 'compatible') {
        return new TerraformCompatibilityResult(
            'Compatible',
            'success',
            'check-square'
        );

    } else if (compatibilityResult == 'incompatible') {
        return new TerraformCompatibilityResult(
            'Incompatible',
            'danger',
            'ban'
        );

    } else if (compatibilityResult == "no_constraint") {
        return new TerraformCompatibilityResult(
            'No version constraint defined',
            'warning',
            'exclamation-triangle'
        );

    } else if (compatibilityResult == "implicit_compatible") {
        return new TerraformCompatibilityResult(
            'Implicitly compatible',
            'primary',
            'check-square'
        );
    }
    return undefined;
}

async function createSearchResultCard(parent_id, type, module) {

    let provider_logos = await getProviderLogos();

    let display_published = timeDifference(new Date(module.published_at));
    let provider_logo_html = '';

    let namespaceDisplayName = module.namespace;
    let namespaceDetails = await getNamespaceDetails(module.namespace);
    if (namespaceDetails.display_name) {
        namespaceDisplayName = namespaceDetails.display_name;
    }

    let link = '';
    if (type == 'module') {
        link = `/modules/${module.namespace}/${module.name}/${module.provider}`;
    } else {
        link = `/providers/${module.namespace}/${module.name}`;
    }

    if (type == "module" && provider_logos[module.provider] !== undefined) {
        let provider_logo_details = provider_logos[module.provider];
        provider_logo_html = `
            <a class="provider-logo-link" href="${provider_logo_details.link}">
                <img style="margin: 5px" height="40" width="40" alt="${provider_logo_details.alt}" src="${provider_logo_details.source}" />
            </a>
        `;
        addProviderLogoTos(module.provider);
    } else if (type == "provider" && module.logo_url) {
        provider_logo_html = `
        <a href="${link}">
            <img style="margin: 5px" height="40" width="40" src="${module.logo_url}" />
        </a>
        `;
    }

    // Replace slashes in ID with full stops
    let card_id = module.id.replace(/\//g, '.');

    let version_compatibility_content = '';

    if (type == "module" && module.version_compatibility) {
        let compatibility_result = getTerraformCompatibilityResultObject(module.version_compatibility);
        if (compatibility_result) {
            version_compatibility_content = `
                <br />
                <p class="card-footer-item card-terraform-version-compatibility">
                    <span class="icon has-text-${compatibility_result.color}">
                        <i class="fas fa-${compatibility_result.icon}"></i>
                    </span>
                    <span>${compatibility_result.text}</span>
                </p>`;
        }
    }

    // Add module to search results
    let result_card = $(
        `
        <div id="${card_id}" class="card">
            <header class="card-header">
                <p class="card-header-title">
                    ${provider_logo_html}
                    <a class="module-card-title" href="${link}">${namespaceDisplayName} / ${module.name}</a>
                </p>
                ${type == "module" ? `
                <a class="module-provider-card-provider-text" href="${link}">
                    <button class="card-header-icon" aria-label="more options">
                        Provider: ${module.provider}
                    </button>
                </a>
                ` : ''}
            </header>
            <a href="${link}">
                <div class="card-content">
                    <div class="content">
                        ${module.description ? module.description : (module.version ? '' : 'This module does not have any published versions')}
                        <br />
                        <br />
                        ${module.owner ? "Owner: " + module.owner : ""}
                    </div>
                </div>
                <footer class="card-footer">
                    <p class="card-footer-item card-source-link">${module.source ? "Source: " + module.source : "No source provided"}</p>
                    <br />
                    <p class="card-footer-item card-last-updated">${module.published_at ? ('Last updated: ' + display_published) : ''}</p>
                    ${version_compatibility_content}
                </footer>
            </a>
        </div>
        <br />
        `
    );
    $(`#${parent_id}`).append(result_card);
    addModuleLabels(module, $(result_card.find('.card-header-title')[0]));
}


terraregModuleDetailsPromiseSingleton = [];

async function getModuleDetails(module_id) {
    // Create promise if it hasn't already been defined
    if (terraregModuleDetailsPromiseSingleton[module_id] === undefined) {
        terraregModuleDetailsPromiseSingleton[module_id] = new Promise((resolve, reject) => {
            let userPreferences = getUserPreferences();
            let terraformVersionConstraintQueryString = (
                userPreferences['terraform-compatibility-version'] ?
                `?target_terraform_version=${userPreferences['terraform-compatibility-version']}` :
                ''
            );

            // Perform request to obtain module details
            $.ajax({
                type: "GET",
                url: `/v1/terrareg/modules/${module_id}${terraformVersionConstraintQueryString}`,
                success: function (data) {
                    resolve(data);
                },
                error: function () {
                    resolve(null);
                }
            });
        });
    }
    return terraregModuleDetailsPromiseSingleton[module_id];
}

terraregModuleUsageBuilderVariablesPromiseSingleton = [];
function getUsageBuilderVariables(moduleId) {
    // Create promise if it hasn't already been defined
    if (terraregModuleUsageBuilderVariablesPromiseSingleton[moduleId] === undefined) {
        terraregModuleUsageBuilderVariablesPromiseSingleton[moduleId] = new Promise((resolve, reject) => {
            // Perform request to obtain usage builder variables details
            $.ajax({
                type: "GET",
                url: `/v1/terrareg/modules/${moduleId}/variable_template`,
                success: function (data) {
                    resolve(data);
                },
                error: function () {
                    resolve(null);
                }
            });
        });
    }
    return terraregModuleUsageBuilderVariablesPromiseSingleton[moduleId];
}


terraregSubmoduleDetailsPromiseSingleton = [];
async function getSubmoduleDetails(moduleId, submodulePath) {
    // Create promise if it hasn't already been defined
    if (terraregSubmoduleDetailsPromiseSingleton[moduleId + submodulePath] === undefined) {
        terraregSubmoduleDetailsPromiseSingleton[moduleId + submodulePath] = new Promise((resolve, reject) => {
            // Perform request to obtain submodule details
            $.ajax({
                type: "GET",
                url: `/v1/terrareg/modules/${moduleId}/submodules/details/${submodulePath}`,
                success: function (data) {
                    resolve(data);
                },
                error: function () {
                    resolve(null);
                }
            });
        });
    }
    return terraregSubmoduleDetailsPromiseSingleton[moduleId + submodulePath];
}

terraregExampleDetailsPromiseSingleton = [];
async function getExampleDetails(moduleId, examplePath) {
    // Create promise if it hasn't already been defined
    if (terraregExampleDetailsPromiseSingleton[moduleId + examplePath] === undefined) {
        terraregExampleDetailsPromiseSingleton[moduleId + examplePath] = new Promise((resolve, reject) => {
            // Perform request to obtain submodule details
            $.ajax({
                type: "GET",
                url: `/v1/terrareg/modules/${moduleId}/examples/details/${examplePath}`,
                success: function (data) {
                    resolve(data);
                },
                error: function () {
                    resolve(null);
                }
            });
        });
    }
    return terraregExampleDetailsPromiseSingleton[moduleId + examplePath];
}

async function addModuleLabels(module, parentDiv) {
    let terrareg_config = await getConfig();
    if (module.trusted) {
        parentDiv.append($(`
            <span class="tag is-info is-light result-card-label result-card-label-trusted">
                <span class="panel-icon">
                    <i class="fas fa-check-circle" aria-hidden="true"></i>
                </span>
                ${terrareg_config.TRUSTED_NAMESPACE_LABEL}
            </span>
        `));
    } else {
        parentDiv.append($(`
            <span class="tag is-warning is-light result-card-label result-card-label-contributed">
                ${terrareg_config.CONTRIBUTED_NAMESPACE_LABEL}
            </span>
        `));
    }

    if (module.verified) {
        parentDiv.append($(`
            <span class="tag is-link is-light result-card-label result-card-label-verified">
                <span class="panel-icon">
                    <i class="fas fa-thumbs-up" aria-hidden="true"></i>
                </span>
                ${terrareg_config.VERIFIED_MODULE_LABEL}
            </span>
        `));
    }

    if (module.internal) {
        parentDiv.append($(`
            <span class="tag is-warning is-light result-card-label result-card-label-internal">
                <span class="panel-icon">
                    <i class="fa fa-eye-slash" aria-hidden="true"></i>
                </span>
                Internal
            </span>
        `));
    }
}
