const form = document.querySelector("#search-form");
const searchTerm = document.querySelector("#search-term");
const chapters = document.querySelectorAll(".chapter");
const searchInput = document.querySelector("#search-term");
const searchType = document.querySelector("#search-by");

form.addEventListener("submit", e => {
  e.preventDefault();

  const term = searchTerm.value.toLowerCase();
  const type = searchType.value;

  const filteredChapters = Array.from(chapters).filter(chapter => {
    const title = chapter.querySelector(".chapter-title").textContent.toLowerCase();
    const description = chapter.querySelector(".chapter-description").textContent.toLowerCase();
    return (type === "title" && title.includes(term)) || (type === "description" && description.includes(term));
  });

  chapters.forEach(chapter => {
    chapter.style.display = "none";
  });

  filteredChapters.forEach(chapter => {
    chapter.style.display = "block";
  });
});
