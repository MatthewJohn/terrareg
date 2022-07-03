function getSetupCardByName(name) {
    return $(`#setup-${name}`);
}
function onSetupCardHeaderClick(event) {
    // From dropdown button, find parent setup card
    let parentCard = $(event.target).closest('.initial-setup-card');
    toggleSetupCard(parentCard);
}
function toggleSetupCard(parentCard) {
    let card = getSetupCardContent(parentCard);
    if (card.is(':visible')) {
        card.hide();
    } else {
        card.show();
    }
}
function getSetupCardContent(cardSelector) {
    return cardSelector.find('.card-content');
}

function strikeThrough(div) {
    div.html('<strike>' + div.html() + '</strike>');
}

function setProgress(percentage) {
    let progressBar = $('#setup-progress-bar');
    progressBar.text(`${percentage}%`);
    progressBar.val(percentage);
}

function isProtocolHttps() {
    return 'https:' == document.location.protocol
}

async function loadSetupPage(overrideHttpsCheck = false) {

    // Hide all cards
    await $('#setup-cards-container').find('.initial-setup-card').each((itx, card) => {
        getSetupCardContent($(card)).hide();
    }).promise();

    let config = await getConfig();
    $.get('/v1/terrareg/initial_setup', async (setupData) => {
        let progressbar = $('setup-progress-bar');

        // Strike through environment variables that have been set
        if (config.ADMIN_AUTHENTICATION_TOKEN_ENABLED) {
            setProgress(10);
            strikeThrough($('#setup-step-auth-vars-admin-authentication-token'));
        }
        if (config.SECRET_KEY_SET) {
            setProgress(10);
            strikeThrough($('#setup-step-auth-vars-secret-key'));
        }
        // If either have not been set, open card and return
        if ((! config.ADMIN_AUTHENTICATION_TOKEN_ENABLED) || (! config.SECRET_KEY_SET)) {
            toggleSetupCard(getSetupCardByName('auth-vars'));
            return;
        }
        setProgress(20);

        // Check if user is logged in
        let loggedIn = await isLoggedIn();
        if (! loggedIn) {
            toggleSetupCard(getSetupCardByName('login'));
            return;
        }
        setProgress(40);

        // Check if module has been created
        if (! setupData.module_created) {
            toggleSetupCard(getSetupCardByName('create-module'));
            return;
        }
        setProgress(60);

        // Populat integration URLs for module's indexing
        $('#module-integrations-link').attr('href', setupData.module_view_url + '#integrations');
        $('.module-upload-endpoint').each((itx, div) => {
            $(div).text(setupData.module_upload_endpoint);
        });
        $('.module-publish-endpoint').each((itx, div) => {
            $(div).text(setupData.module_publish_endpoint);
        });

        // Check if module version has been indexed
        if (!setupData.version_indexed || !setupData.version_published) {

            // Check if version has been indexed, but has not been published
            if (setupData.version_indexed) {
                setProgress(70);
                // Strike through the upload steps
                $('.setup-step-upload-module-version').each((itx, div) => {
                    strikeThrough($(div));
                })
                // Display warning in git upload
                $('#setup-step-index-git-not-published-warning').removeClass('default-hidden');
            }

            // If git is configured, show this card
            if (setupData.module_configured_with_git) {
                toggleSetupCard(getSetupCardByName('index-git'));
                return;
            } else {
                toggleSetupCard(getSetupCardByName('index-upload'));
                return;
            }
        }
        setProgress(80);

        let secureTasksRemaining = 0;
        // Check upload API keys
        if (config.UPLOAD_API_KEYS_ENABLED || ! config.ALLOW_MODULE_HOSTING) {
            strikeThrough($('#setup-step-secure-upload'));
        } else {
            secureTasksRemaining += 1;
        }

        // Check publish API keys
        if (config.PUBLISH_API_KEYS_ENABLED) {
            strikeThrough($('#setup-step-secure-publish'));
        } else {
            secureTasksRemaining += 1;
        }

        if (secureTasksRemaining) {
            setProgress(100 - (secureTasksRemaining * 10));
            toggleSetupCard(getSetupCardByName('secure'));
            return;
        }

        // Check if URL is HTTPs
        if (! isProtocolHttps() && !overrideHttpsCheck) {
            toggleSetupCard(getSetupCardByName('ssl'));
            setProgress(100);
            return;
        }

        // Display complete
        setProgress(120);
        toggleSetupCard(getSetupCardByName('complete'));
    });
}
