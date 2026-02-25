// Carousel functionality
document.addEventListener('DOMContentLoaded', function() {
    const carouselTrack = document.getElementById('carouselTrack');
    const prevBtn = document.querySelector('.carousel-btn-prev');
    const nextBtn = document.querySelector('.carousel-btn-next');
    
    if (!carouselTrack || !prevBtn || !nextBtn) return;
    
    let currentIndex = 0;
    const items = carouselTrack.querySelectorAll('.carousel-item');
    const itemCount = items.length;
    const visibleItems = 3;
    
    // Update carousel position
    function updateCarousel() {
        // Calculate the width of one item (including gap)
        const itemWidth = 100 / visibleItems; // percentage
        const translateX = -currentIndex * itemWidth;
        carouselTrack.style.transform = `translateX(${translateX}%)`;
    }
    
    // Next button
    nextBtn.addEventListener('click', function() {
        if (currentIndex < itemCount - visibleItems) {
            currentIndex++;
            updateCarousel();
        }
    });
    
    // Previous button
    prevBtn.addEventListener('click', function() {
        if (currentIndex > 0) {
            currentIndex--;
            updateCarousel();
        }
    });
    
    // Keyboard navigation (optional)
    document.addEventListener('keydown', function(e) {
        if (e.key === 'ArrowRight') nextBtn.click();
        if (e.key === 'ArrowLeft') prevBtn.click();
    });
});
