// Minimal ScrollSpy fix for fixed sidebar positioning
document.addEventListener("DOMContentLoaded", function () {
  const sidebar = document.querySelector(".docs-toc");
  const tocNav = document.querySelector("#toc");

  if (!sidebar || !tocNav) return;

  // Function to update active states manually since Bootstrap ScrollSpy
  // doesn't work well with fixed positioning
  function updateActiveStates() {
    const sections = document.querySelectorAll(
      "h1[id], h2[id], h3[id], h4[id], h5[id], h6[id]",
    );
    const tocLinks = tocNav.querySelectorAll('a[href^="#"]');

    let activeSection = null;
    const scrollPos = window.scrollY + 100; // Offset for better detection

    // Find the current section
    sections.forEach((section) => {
      if (section.offsetTop <= scrollPos) {
        activeSection = section;
      }
    });

    // Remove all active states
    tocLinks.forEach((link) => {
      link.classList.remove("active");
      const li = link.closest("li");
      if (li) li.classList.remove("active");
    });

    // Add active state to current section
    if (activeSection) {
      const activeLink = tocNav.querySelector(`a[href="#${activeSection.id}"]`);
      if (activeLink) {
        activeLink.classList.add("active");
        const li = activeLink.closest("li");
        if (li) li.classList.add("active");

        // Auto-scroll sidebar to keep active section visible
        const sidebar = activeLink.closest(".docs-toc");
        if (sidebar) {
          const linkRect = activeLink.getBoundingClientRect();
          const sidebarRect = sidebar.getBoundingClientRect();

          // Check if link is outside visible area
          if (
            linkRect.top < sidebarRect.top + 50 ||
            linkRect.bottom > sidebarRect.bottom - 50
          ) {
            // Scroll the active link into view within the sidebar
            activeLink.scrollIntoView({
              behavior: "smooth",
              block: "center",
              inline: "nearest",
            });
          }
        }
      }
    }
  }

  // Throttled scroll handler for performance
  let ticking = false;
  function onScroll() {
    if (!ticking) {
      requestAnimationFrame(() => {
        updateActiveStates();
        ticking = false;
      });
      ticking = true;
    }
  }

  // Add smooth scrolling to TOC links
  tocNav.addEventListener("click", function (e) {
    if (
      e.target.tagName === "A" &&
      e.target.getAttribute("href").startsWith("#")
    ) {
      e.preventDefault();
      const targetId = e.target.getAttribute("href").substring(1);
      const targetElement = document.getElementById(targetId);

      if (targetElement) {
        targetElement.scrollIntoView({
          behavior: "smooth",
          block: "start",
        });
      }
    }
  });

  // Initialize
  window.addEventListener("scroll", onScroll, { passive: true });
  updateActiveStates(); // Set initial state
});
