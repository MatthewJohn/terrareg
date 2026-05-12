/**
 * Provider Sources Page JavaScript
 * Handles the display and management of provider sources
 */

class ProviderSourcesPage {
    constructor() {
        this.providerSources = [];
        this.currentDeleteSource = null;
        this.init();
    }

    init() {
        this.bindEvents();
        this.loadProviderSources();
    }

    bindEvents() {
        // Delete button clicks
        $(document).on('click', '.delete-provider-source-button', (e) => {
            const providerSourceName = $(e.currentTarget).data('name');
            this.showDeleteModal(providerSourceName);
        });

        // Cancel delete
        $('#cancel-delete-button').on('click', () => {
            this.hideDeleteModal();
        });

        // Confirm delete
        $('#confirm-delete-button').on('click', () => {
            if (this.currentDeleteSource) {
                this.deleteProviderSource(this.currentDeleteSource);
            }
        });

        // Close modal on background click
        $('.modal-background, .modal-card-head .delete').on('click', () => {
            this.hideDeleteModal();
        });

        // Close error/success messages
        $('.notification .delete').on('click', () => {
            $(this).parent().addClass('default-hidden');
        });
    }

    loadProviderSources() {
        $.get('/v1/terrareg/provider-sources')
            .done((data) => {
                this.providerSources = data;
                this.renderTable();
            })
            .fail((xhr) => {
                this.showError('Failed to load provider sources: ' + xhr.responseJSON?.message || 'Unknown error');
            });
    }

    renderTable() {
        const tbody = $('#provider-sources-table-body');
        tbody.empty();

        if (this.providerSources.length === 0) {
            tbody.append('<tr><td colspan="4" class="has-text-centered">No provider sources found</td></tr>');
            return;
        }

        this.providerSources.forEach(source => {
            const row = $(`
                <tr>
                    <td>${this.escapeHtml(source.name)}</td>
                    <td>${this.escapeHtml(source.provider_source_type)}</td>
                    <td>${this.escapeHtml(source.api_name)}</td>
                    <td>
                        <button class="button is-danger is-small delete-provider-source-button" data-name="${this.escapeHtml(source.name)}">
                            <span class="icon">
                                <i class="fas fa-trash"></i>
                            </span>
                            <span>Delete</span>
                        </button>
                    </td>
                </tr>
            `);
            tbody.append(row);
        });
    }

    showDeleteModal(providerSourceName) {
        this.currentDeleteSource = providerSourceName;
        $('#delete-provider-source-name').text(providerSourceName);
        $('#delete-modal').removeClass('default-hidden');
    }

    hideDeleteModal() {
        $('#delete-modal').addClass('default-hidden');
        this.currentDeleteSource = null;
    }

    deleteProviderSource(providerSourceName) {
        $.ajax({
            url: `/v1/terrareg/provider-sources/${encodeURIComponent(providerSourceName)}`,
            method: 'DELETE',
            headers: {
                'X-CSRFToken': this.getCsrfToken()
            }
        })
        .done(() => {
            this.hideDeleteModal();
            this.showSuccess(`Provider source "${providerSourceName}" deleted successfully`);
            this.loadProviderSources();
        })
        .fail((xhr) => {
            const errorMessage = xhr.responseJSON?.message || 'Failed to delete provider source';
            this.showError(errorMessage);
            this.hideDeleteModal();
        });
    }

    showError(message) {
        $('#error-message-text').text(message);
        $('#error-message').removeClass('default-hidden');
        $(window).scrollTop($('#error-message').offset().top);
    }

    showSuccess(message) {
        $('#success-message-text').text(message);
        $('#success-message').removeClass('default-hidden');
        $(window).scrollTop($('#success-message').offset().top);
    }

    getCsrfToken() {
        // Get CSRF token from meta tag or cookie
        const metaTag = document.querySelector('meta[name="csrf-token"]');
        if (metaTag) {
            return metaTag.getAttribute('content');
        }
        return '';
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

function renderPage() {
    window.providerSourcesPage = new ProviderSourcesPage();
}
