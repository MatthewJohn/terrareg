
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

async function createSearchResultCard(parent_id, module) {

    let provider_logos = await getProviderLogos();

    let display_published = timeDifference(new Date(module.published_at));
    let provider_logo_html = '';
    if (provider_logos[module.provider] !== undefined) {
        let provider_logo_details = provider_logos[module.provider];
        provider_logo_html = `
            <a href="${provider_logo_details.link}">
                <img style="margin: 5px" height="40" width="40" alt="${provider_logo_details.alt}" src="${provider_logo_details.source}" />
            </a>
        `;

        // Add provider TOS to results, if not already there
        if ($('#provider-tos-' + module.provider).length == 0) {
            let tos_object = document.createElement('p');
            tos_object.id = `provider-tos-${module.provider}`;
            tos_object.innerHTML = provider_logo_details.tos;
            $('#provider-tos')[0].append(tos_object);
        }
    }

    // Replace slashes in ID with full stops
    let card_id = module.id.replace(/\//g, '.');

    // Add module to search results
    let result_card = $(
        `  
        <div id="${card_id}" class="card">
            <header class="card-header">
                <p class="card-header-title">
                    ${provider_logo_html}
                    <a class="module-card-title" href="/modules/${module.id}">${module.namespace} / ${module.name}</a>
                </p>
                <a href="/modules/${module.id}">
                    <button class="card-header-icon" aria-label="more options">
                        <span class="icon">
                            <i class="fas fa-external-link" aria-hidden="true"></i>
                        </span>
                    </button>
                </a>
            </header>
            <a href="/modules/${module.id}">
                <div class="card-content">
                    <div class="content">
                        ${module.description ? "Description<br />" + module.description : ""}
                        <br />
                        <br />
                        ${module.owner ? "Owner: " + module.owner : ""}
                    </div>
                </div>
                <footer class="card-footer">
                    <p class="card-footer-item card-source-link">${module.source? "Source: " + module.source : "No source provided"}</p>
                    <br />
                    <p class="card-footer-item card-last-updated">Last updated: ${display_published}</p>
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
            // Perform request to obtain module details
            $.ajax({
                type: "GET",
                url: `/v1/modules/${module_id}`,
                success: function (data) {
                    // append module provider ID to data
                    data.module_provider_id = data.id.split('/').slice(0, 3).join('/');
                    resolve(data);
                }
            });
        });
    }
    return terraregModuleDetailsPromiseSingleton[module_id];
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
