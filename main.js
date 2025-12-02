document.addEventListener('DOMContentLoaded', () => {
    // --- DOM Element Selectors ---
    const sidebar = document.getElementById('sidebar');
    const menuToggle = document.getElementById('menu-toggle');
    const themeToggleDesktop = document.getElementById('theme-toggle-desktop');
    const themeToggleMobile = document.getElementById('theme-toggle-mobile');
    const searchInput = document.getElementById('search-input');
    const tocContainer = document.getElementById('toc');
    const content = document.getElementById('documentation-content');

    // --- Initialize All Features ---
    if (sidebar && menuToggle) {
        initMobileSidebar();
    }
    if (themeToggleDesktop && themeToggleMobile) {
        initThemeToggle();
    }
    if (content) {
        initCopyToClipboard();
        if (tocContainer) {
            initScrollspy();
        }
        if (searchInput) {
            initSearch();
        }
        initTabs();
        initCollapsibleSections();
    }

    // --- Feature Implementations ---

    /**
     * Handles the mobile sidebar opening and closing.
     */
    function initMobileSidebar() {
        menuToggle.addEventListener('click', () => {
            sidebar.classList.toggle('-translate-x-full');
        });
    }

    /**
     * Handles the light/dark mode theme toggle.
     */
    function initThemeToggle() {
        const sunIcon = `<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M12 6a6 6 0 100 12 6 6 0 000-12z"></path></svg>`;
        const moonIcon = `<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"></path></svg>`;

        const updateThemeIcons = () => {
            const isDark = document.documentElement.classList.contains('dark');
            const icon = isDark ? sunIcon : moonIcon;
            themeToggleDesktop.innerHTML = icon;
            themeToggleMobile.innerHTML = icon;
        };

        const toggleTheme = () => {
            const isDark = document.documentElement.classList.toggle('dark');
            localStorage.setItem('theme', isDark ? 'dark' : 'light');
            updateThemeIcons();
        };

        themeToggleDesktop.addEventListener('click', toggleTheme);
        themeToggleMobile.addEventListener('click', toggleTheme);
        updateThemeIcons(); // Set initial icon
    }

    /**
     * Generates the Table of Contents and handles active link highlighting on scroll.
     */
    function initScrollspy() {
        const sections = Array.from(content.querySelectorAll('section[id]'));
        const tocLinks = [];
        const tocList = document.createElement('ul');
        tocList.className = 'space-y-1';

        sections.forEach(section => {
            const h2 = section.querySelector('h2');
            const titleEl = section.id === 'overview' ? section.querySelector('h1') : h2;
            if (!titleEl) return;

            const title = titleEl.textContent;
            const li = document.createElement('li');
            const a = document.createElement('a');

            a.href = `#${section.id}`;
            a.textContent = title;
            a.className = 'toc-link block py-2 px-3 rounded-md border-l-2 border-transparent hover:bg-gray-200 dark:hover:bg-gray-800 hover:border-gray-400 dark:hover:border-gray-600 transition-colors duration-200';
            a.dataset.sectionId = section.id;

            a.addEventListener('click', (e) => {
                e.preventDefault();
                section.scrollIntoView({ behavior: 'smooth' });
                if (window.innerWidth < 768) {
                    sidebar.classList.add('-translate-x-full');
                }
            });

            li.appendChild(a);
            tocList.appendChild(li);
            tocLinks.push(a);
        });

        tocContainer.appendChild(tocList);

        const observer = new IntersectionObserver(entries => {
            let latestIntersecting = null;
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    if (!latestIntersecting || entry.boundingClientRect.top < latestIntersecting.boundingClientRect.top) {
                        latestIntersecting = entry;
                    }
                }
            });

            tocLinks.forEach(l => l.classList.remove('active', 'text-blue-500', 'border-blue-500'));
            if (latestIntersecting) {
                const id = latestIntersecting.target.getAttribute('id');
                const activeLink = tocContainer.querySelector(`a[data-section-id="${id}"]`);
                if (activeLink) {
                    activeLink.classList.add('active', 'text-blue-500', 'border-blue-500');
                }
            }
        }, { rootMargin: '-50px 0px -50% 0px', threshold: 0 });

        sections.forEach(section => observer.observe(section));
    }

    /**
     * Adds copy-to-clipboard functionality to all code blocks.
     */
    function initCopyToClipboard() {
        content.querySelectorAll('.copy-btn').forEach(copyButton => {
            const codeBlock = copyButton.closest('.code-block');
            const code = codeBlock.querySelector('pre code');

            if (!code) return;

            copyButton.addEventListener('click', () => {
                navigator.clipboard.writeText(code.innerText).then(() => {
                    copyButton.textContent = 'Copied!';
                    setTimeout(() => {
                        copyButton.textContent = 'Copy';
                    }, 2000);
                }).catch(err => {
                    copyButton.textContent = 'Failed!';
                    console.error('Failed to copy text: ', err);
                });
            });
        });
    }

    /**
     * Implements client-side search functionality.
     */
    function initSearch() {
        const sections = Array.from(content.querySelectorAll('section[id]'));
        searchInput.addEventListener('input', (e) => {
            const searchTerm = e.target.value.toLowerCase();
            sections.forEach(section => {
                const title = (section.querySelector('h2')?.textContent || section.querySelector('h1')?.textContent || '').toLowerCase();
                const body = section.innerText.toLowerCase();
                const isMatch = title.includes(searchTerm) || body.includes(searchTerm);
                section.style.display = isMatch ? '' : 'none';
            });
        });
    }

    /**
     * Handles tab switching for code examples or other tabbed content.
     */
    function initTabs() {
        document.querySelectorAll('.tabs-container').forEach(tabContainer => {
            const tabButtons = tabContainer.querySelectorAll('.tab-btn');
            const tabContents = tabContainer.querySelectorAll('.tab-content');

            tabButtons.forEach(button => {
                button.addEventListener('click', () => {
                    const tabName = button.dataset.tab;

                    tabButtons.forEach(btn => {
                        btn.classList.remove('active');
                        if (btn === button) {
                            btn.classList.add('active');
                        }
                    });

                    tabContents.forEach(content => {
                        content.classList.remove('active');
                        if (content.dataset.tabContent === tabName) {
                            content.classList.add('active');
                        }
                    });
                });
            });
        });
    }
    
    /**
     * Makes sections collapsible.
     */
    function initCollapsibleSections() {
        content.querySelectorAll('h2').forEach(h2 => {
            const section = h2.closest('section');
            if (!section || section.id === 'overview') return;

            h2.classList.add('cursor-pointer', 'flex', 'justify-between', 'items-center');
            
            const chevron = document.createElement('span');
            chevron.className = 'transform transition-transform duration-200';
            chevron.innerHTML = '▼';
            h2.appendChild(chevron);

            const collapsibleContent = Array.from(section.children).filter(child => child !== h2);

            const toggleCollapse = () => {
                const isCollapsed = section.classList.toggle('collapsed');
                collapsibleContent.forEach(el => {
                    el.style.display = isCollapsed ? 'none' : '';
                });
                chevron.style.transform = isCollapsed ? 'rotate(-90deg)' : 'rotate(0deg)';
            };

            h2.addEventListener('click', toggleCollapse);

            // You can uncomment this if you want sections collapsed by default:
            // if (section.id !== 'quickstart' && section.id !== 'beginner-guide') {
            //     toggleCollapse();
            // }
        });
    }
});
