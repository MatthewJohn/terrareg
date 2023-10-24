
// async function createSearchResultCard(parent_id, module) {

//     let provider_logos = await getProviderLogos();

//     let display_published = timeDifference(new Date(module.published_at));
//     let provider_logo_html = '';

//     let namespaceDisplayName = module.namespace;
//     let namespaceDetails = await getNamespaceDetails(module.namespace);
//     if (namespaceDetails.display_name) {
//         namespaceDisplayName = namespaceDetails.display_name;
//     }

//     if (provider_logos[module.provider] !== undefined) {
//         let provider_logo_details = provider_logos[module.provider];
//         provider_logo_html = `
//             <a href="${provider_logo_details.link}">
//                 <img style="margin: 5px" height="40" width="40" alt="${provider_logo_details.alt}" src="${provider_logo_details.source}" />
//             </a>
//         `;
//         addProviderLogoTos(module.provider);
//     }

//     // Replace slashes in ID with full stops
//     let card_id = module.id.replace(/\//g, '.');

//     let link = `/modules/${module.namespace}/${module.name}/${module.provider}`;

//     let version_compatibility_content = '';

//     if (module.version_compatibility) {
//         let compatibility_result = getTerraformCompatibilityResultObject(module.version_compatibility);
//         if (compatibility_result) {
//             version_compatibility_content = `
//                 <br />
//                 <p class="card-footer-item card-terraform-version-compatibility">
//                     <span class="icon has-text-${compatibility_result.color}">
//                         <i class="fas fa-${compatibility_result.icon}"></i>
//                     </span>
//                     <span>${compatibility_result.text}</span>
//                 </p>`;
//         }
//     }

//     // Add module to search results
//     let result_card = $(
//         `
//         <div id="${card_id}" class="card">
//             <header class="card-header">
//                 <p class="card-header-title">
//                     ${provider_logo_html}
//                     <a class="module-card-title" href="${link}">${namespaceDisplayName} / ${module.name}</a>
//                 </p>
//                 <a class="module-provider-card-provider-text" href="${link}">
//                     <button class="card-header-icon" aria-label="more options">
//                         Provider: ${module.provider}
//                     </button>
//                 </a>
//             </header>
//             <a href="${link}">
//                 <div class="card-content">
//                     <div class="content">
//                         ${module.description ? module.description : (module.version ? '' : 'This module does not have any published versions')}
//                         <br />
//                         <br />
//                         ${module.owner ? "Owner: " + module.owner : ""}
//                     </div>
//                 </div>
//                 <footer class="card-footer">
//                     <p class="card-footer-item card-source-link">${module.source ? "Source: " + module.source : "No source provided"}</p>
//                     <br />
//                     <p class="card-footer-item card-last-updated">${module.published_at ? ('Last updated: ' + display_published) : ''}</p>
//                     ${version_compatibility_content}
//                 </footer>
//             </a>
//         </div>
//         <br />
//         `
//     );
//     $(`#${parent_id}`).append(result_card);
//     addModuleLabels(module, $(result_card.find('.card-header-title')[0]));
// }


terraregProviderDetailsPromiseSingleton = [];
async function getProviderDetails(providerId) {
    // Create promise if it hasn't already been defined
    if (terraregProviderDetailsPromiseSingleton[providerId] === undefined) {
        terraregProviderDetailsPromiseSingleton[providerId] = new Promise((resolve, reject) => {
            // Perform request to obtain module details
            $.ajax({
                type: "GET",
                url: `/v1/providers/${providerId}`,
                success: function (data) {
                    console.log(data);
                    resolve(data);
                },
                error: function () {
                    resolve(null);
                }
            });
        });
    }
    return terraregProviderDetailsPromiseSingleton[providerId];
}

// async function addProviderLabels(provider, parentDiv) {
//     let terrareg_config = await getConfig();
//     if (provider.trusted) {
//         parentDiv.append($(`
//             <span class="tag is-info is-light result-card-label result-card-label-trusted">
//                 <span class="panel-icon">
//                     <i class="fas fa-check-circle" aria-hidden="true"></i>
//                 </span>
//                 ${terrareg_config.TRUSTED_NAMESPACE_LABEL}
//             </span>
//         `));
//     } else {
//         parentDiv.append($(`
//             <span class="tag is-warning is-light result-card-label result-card-label-contributed">
//                 ${terrareg_config.CONTRIBUTED_NAMESPACE_LABEL}
//             </span>
//         `));
//     }

//     if (provider.verified) {
//         parentDiv.append($(`
//             <span class="tag is-link is-light result-card-label result-card-label-verified">
//                 <span class="panel-icon">
//                     <i class="fas fa-thumbs-up" aria-hidden="true"></i>
//                 </span>
//                 ${terrareg_config.VERIFIED_MODULE_LABEL}
//             </span>
//         `));
//     }

//     if (provider.internal) {
//         parentDiv.append($(`
//             <span class="tag is-warning is-light result-card-label result-card-label-internal">
//                 <span class="panel-icon">
//                     <i class="fa fa-eye-slash" aria-hidden="true"></i>
//                 </span>
//                 Internal
//             </span>
//         `));
//     }
// }
