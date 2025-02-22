document.addEventListener('DOMContentLoaded', function () {
    const snowflakesContainer = document.getElementById('snowflakes-container');

    function createSnowflake() {
        const snowflake = document.createElement('div');
        snowflake.classList.add('snowflake');
        snowflake.textContent = '❄';
        snowflake.style.left = Math.random() * 100 + '%';
        snowflake.style.animationDuration = Math.random() * 3 + 5 + 's';
        snowflake.style.fontSize = Math.random() * 50 + 10 + 'px';

        snowflakesContainer?.appendChild(snowflake);

        setTimeout(() => {
            snowflake.remove();
        }, parseFloat(snowflake.style.animationDuration) * 1000);
    }

    setInterval(createSnowflake, 100);
});