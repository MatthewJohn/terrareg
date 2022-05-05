

function timeDifference(previous) {
    var msPerMinute = 60 * 1000;
    var msPerHour = msPerMinute * 60;
    var msPerDay = msPerHour * 24;
    var msPerMonth = msPerDay * 30;
    var msPerYear = msPerDay * 365;

    var current = new Date();
    var elapsed = current - previous;

    if (elapsed < msPerMinute) {
        let cnt = Math.round(elapsed / 1000);
        if (cnt == 1) {
            return cnt + ' second ago';
        }
        return cnt + ' seconds ago';
    }

    else if (elapsed < msPerHour) {
        let cnt = Math.round(elapsed / msPerMinute);
        if (cnt == 1) {
            return cnt + ' minute ago';
        }
        return cnt + ' minutes ago';
    }

    else if (elapsed < msPerDay ) {
        let cnt = Math.round(elapsed / msPerHour);
        if (cnt == 1) {
            return cnt + ' hour ago';
        }
        return cnt + ' hours ago';
    }

    else if (elapsed < msPerMonth) {
        let cnt = Math.round(elapsed / msPerDay);
        if (cnt == 1) {
            return 'approximately ' + cnt + ' day ago'
        }
        return 'approximately ' + cnt + ' days ago';
    }

    else if (elapsed < msPerYear) {
        let cnt = Math.round(elapsed / msPerMonth);
        if (cnt == 1) {
            return 'approximately ' + cnt + ' month ago';
        }
        return 'approximately ' + cnt + ' months ago';
    }

    else {
        let cnt = Math.round(elapsed / msPerYear);
        if (cnt == 1) {
            return 'approximately ' + cnt + ' year ago';
        }
        return 'approximately ' + cnt + ' years ago';   
    }
}

function createSearchResultCard(parent_id, module, provider_logos) {
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
            console.log('ADding now');
            console.log($('#provider-tos')[0]);
            let tos_object = document.createElement('p');
            tos_object.id = `provider-tos-${module.provider}`;
            tos_object.innerHTML = provider_logo_details.tos;
            $('#provider-tos')[0].append(tos_object);
        }
    }

    // Add module to search results
    $(`#${parent_id}`).append(
        `  
        <div class="card">
            <header class="card-header">
                <p class="card-header-title">
                    ${provider_logo_html}
                    <a href="/modules/${module.id}">${module.namespace} / ${module.name}</a>
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
                        ${module.description ? "Description<br />" + module.description : "No description provided"}
                        <br />
                        <br />
                        ${module.owner ? "Owner: " + module.owner : ""}
                    </div>
                </div>
                <footer class="card-footer">
                    <p class="card-footer-item">${module.source? "Source: " + module.source : "No source provided"}</p>
                    <br />
                    <p class="card-footer-item">Last updated: ${display_published}</p>
                </footer>
            </a>
        </div>
        <br />
        `
    );
}
