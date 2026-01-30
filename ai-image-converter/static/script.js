document.addEventListener('DOMContentLoaded', () => {
    const tabs = document.querySelectorAll('.tab');
    const sidebarItems = document.querySelectorAll('aside li');
    const homeContent = document.getElementById('home-content');
    const categoryContent = document.getElementById('category-content');
    const categoryTitle = document.getElementById('category-title');
    const categoryDesc = document.getElementById('category-desc');
    const categoryInput = document.getElementById('category-input');
    const convertForm = document.getElementById('convert-form');
    const result = document.getElementById('result');
    const resultImg = document.getElementById('result-img');
    const download = document.getElementById('download');
    const beforeImg = document.getElementById('before-img');
    const afterImg = document.getElementById('after-img');
    const processing = document.getElementById('processing');
    const convertBtn = document.getElementById('convert-btn');
    const homeConvertForm = document.getElementById('home-convert-form');
    const homeUpload = document.getElementById('home-upload');
    const homeFileName = document.getElementById('home-file-name');
    const heroConvertBtn = document.querySelector('.hero-convert-btn');

    const categories = {
        bw: { title: 'Black & White', desc: 'Convert your image to grayscale.' },
        cartoon: { title: 'Cartoon Converter', desc: 'Select a cartoon style and convert your image.' },
        removebg: { title: 'Background Removal', desc: 'Automatically remove the background from your image.' },
        changebg: { title: 'Background Change', desc: 'Select a new background and apply it to your image.' }
    };

    // Card button clicks
    document.querySelectorAll('.card button').forEach(btn => {
        btn.addEventListener('click', () => {
            const cat = btn.closest('.card').dataset.category;
            showCategory(cat);
        });
    });

    function validateFile(file) {
        const allowedTypes = ['image/jpeg', 'image/png'];
        const maxSize = 10 * 1024 * 1024; // 10MB
        if (!allowedTypes.includes(file.type)) {
            alert('Please select a valid image file (JPG, PNG).');
            return false;
        }
        if (file.size > maxSize) {
            alert('File size must be less than 10MB.');
            return false;
        }
        return true;
    }

    // Tab clicks
    tabs.forEach(tab => {
        tab.addEventListener('click', (e) => {
            e.preventDefault();
            const tabName = tab.dataset.tab;
            if (tabName === 'home') {
                showHome();
            } else {
                showCategory(tabName);
            }
        });
    });

    // Sidebar clicks
    sidebarItems.forEach(item => {
        item.addEventListener('click', () => {
            const cat = item.dataset.category;
            showCategory(cat);
        });
    });

    // File input change for category
    const fileInput = document.getElementById('image-upload');
    const fileName = document.getElementById('file-name');
    fileInput.addEventListener('change', (e) => {
        handleFileChange(e.target.files, fileName, beforeImg, convertBtn);
    });

    // File input change for home
    homeUpload.addEventListener('change', (e) => {
        handleFileChange(e.target.files, homeFileName, null, heroConvertBtn);
    });

    // Drag and drop for home
    const uploadArea = document.getElementById('upload-area');
    uploadArea.addEventListener('dragover', (e) => {
        e.preventDefault();
        uploadArea.classList.add('dragover');
    });
    uploadArea.addEventListener('dragleave', () => {
        uploadArea.classList.remove('dragover');
    });
    uploadArea.addEventListener('drop', (e) => {
        e.preventDefault();
        uploadArea.classList.remove('dragover');
        const files = e.dataTransfer.files;
        if (files.length > 0) {
            homeUpload.files = files;
            handleFileChange(files, homeFileName, null, heroConvertBtn);
        }
    });

    function handleFileChange(files, nameSpan, previewImg, btn) {
        if (files.length > 0) {
            const file = files[0];
            if (!validateFile(file)) {
                nameSpan.textContent = 'No file chosen';
                if (btn) btn.disabled = true;
                return;
            }
            nameSpan.textContent = file.name;
            if (previewImg) {
                previewImg.src = URL.createObjectURL(file);
                previewImg.parentElement.classList.add('uploaded');
            }
            if (btn) btn.disabled = false;
        } else {
            nameSpan.textContent = 'No file chosen';
            if (btn) btn.disabled = true;
        }
    }

    // Form submit for category
    convertForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        await processConversion(new FormData(convertForm), afterImg, resultImg, download, result);
    });

    // Form submit for home
    homeConvertForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        await processConversion(new FormData(homeConvertForm), null, null, null, null);
        alert('Image converted successfully! Check the B&W category for more options.');
    });

    async function processConversion(formData, afterImg, resultImg, download, result) {
        processing.style.display = 'block';
        convertBtn.disabled = true;
        heroConvertBtn.disabled = true;
        try {
            const response = await fetch('/convert', {
                method: 'POST',
                body: formData
            });
            const data = await response.json();
            if (response.ok) {
                if (afterImg) {
                    afterImg.src = data.image;
                    afterImg.parentElement.classList.add('uploaded');
                }
                if (resultImg) resultImg.src = data.image;
                if (download) download.href = data.image;
                if (result) result.style.display = 'block';
                updateRemaining(data.remaining);
            } else {
                alert(data.error || 'Error');
            }
        } catch (err) {
            alert('Network error or invalid response');
        } finally {
            processing.style.display = 'none';
            convertBtn.disabled = false;
            heroConvertBtn.disabled = false;
        }
    }

    function showHome() {
        homeContent.style.display = 'block';
        categoryContent.style.display = 'none';
        tabs.forEach(t => t.classList.remove('active'));
        document.querySelector('.tab[data-tab="home"]').classList.add('active');
    }

    function showCategory(cat) {
        homeContent.style.display = 'none';
        categoryContent.style.display = 'block';
        categoryTitle.textContent = categories[cat].title;
        categoryDesc.textContent = categories[cat].desc;
        categoryInput.value = cat;
        result.style.display = 'none';
        processing.style.display = 'none';
        // Reset images: use real photo for 'before' and local sample for 'after'
        beforeImg.src = getRemoteSample(cat);
        afterImg.src = `/sample?category=${cat}&type=after`;
        beforeImg.parentElement.classList.add('uploaded');
        afterImg.parentElement.classList.remove('uploaded');

        function getRemoteSample(c) {
            switch (c) {
                case 'bw': return '/static/images/sample_bw.svg?v=1';
                case 'cartoon': return '/static/images/sample_cartoon.svg?v=1';
                case 'removebg': return '/static/images/sample_removebg.svg?v=1';
                case 'changebg': return '/static/images/sample_changebg.svg?v=1';
                default: return '/static/images/sample_changebg.svg?v=1';
            }
        }
        // Reset file input
        fileInput.value = '';
        fileName.textContent = 'No file chosen';
        convertBtn.disabled = true;
        const styleSelect = document.getElementById('style-select');
        styleSelect.innerHTML = '<option value="">Select an option</option>';
        if (cat === 'cartoon') {
            styleSelect.style.display = 'block';
            styleSelect.innerHTML += '<option value="classic">Classic Cartoon</option><option value="modern">Modern Cartoon</option><option value="anime">Anime Style</option>';
            styleSelect.required = true;
        } else if (cat === 'changebg') {
            styleSelect.style.display = 'block';
            styleSelect.innerHTML += '<option value="blue">Blue Background</option><option value="red">Red Background</option><option value="beach">Beach Scene</option><option value="forest">Forest</option>';
            styleSelect.required = true;
        } else {
            styleSelect.style.display = 'none';
            styleSelect.required = false;
        }
        // Update active tab and sidebar
        tabs.forEach(t => t.classList.remove('active'));
        document.querySelector(`.tab[data-tab="${cat}"]`).classList.add('active');
        sidebarItems.forEach(i => i.classList.remove('active'));
        document.querySelector(`aside li[data-category="${cat}"]`).classList.add('active');
    }

    function updateRemaining(rem) {
        document.getElementById('remaining').textContent = `${rem} conversions left`;
    }

    // Load remaining from cookie
    const remCookie = document.cookie.split(';').find(c => c.trim().startsWith('remaining='));
    if (remCookie) {
        const rem = remCookie.split('=')[1];
        updateRemaining(rem);
    }

    // Slider functionality
    const slides = document.querySelectorAll('.slide');
    const prevBtn = document.querySelector('.slider-btn.prev');
    const nextBtn = document.querySelector('.slider-btn.next');
    let currentSlide = 0;

    function showSlide(index) {
        slides.forEach((slide, i) => {
            slide.classList.remove('active');
            if (i === index) slide.classList.add('active');
        });
    }

    prevBtn.addEventListener('click', () => {
        currentSlide = (currentSlide - 1 + slides.length) % slides.length;
        showSlide(currentSlide);
    });

    nextBtn.addEventListener('click', () => {
        currentSlide = (currentSlide + 1) % slides.length;
        showSlide(currentSlide);
    });

    showSlide(0); // Initial show

    // Auto slide
    setInterval(() => {
        currentSlide = (currentSlide + 1) % slides.length;
        showSlide(currentSlide);
    }, 3000);

    // Default to home
    showHome();
    heroConvertBtn.disabled = true;
});