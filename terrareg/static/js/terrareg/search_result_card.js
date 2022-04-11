

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

function createSearchResultCard(parent_id, module) {
    let display_published = timeDifference(new Date(module.published_at));
    // Add module to search results
    $(`#${parent_id}`).append(
        `
        <a href="/modules/${module.id}">
            <div class="card">
                <header class="card-header">
                    <p class="card-header-title">
                        ${module.namespace} / ${module.name}
                    </p>
                    <button class="card-header-icon" aria-label="more options">
                    <span class="icon">
                        <i class="fas fa-external-link" aria-hidden="true"></i>
                    </span>
                    </button>
                </header>
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
            </div>
        </a>
        <br />
        `
    );
}
