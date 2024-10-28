/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ['./**/*.{html,templ,go}'],
    theme: {
        colors: {
            'base': '#fafaf8',
            'content': '#141413',
            'accent': '#f97316',
            'blue': '#0064E6',
        },

        fontSize: {
            // @link https://utopia.fyi/type/calculator?c=320,16,1.2,1240,22,1.25,5,2,&s=0.75|0.5|0.25,1.5|2|3|4|6|8|11|15|19,3xs-xs|2xs-s|xs-m|s-l|m-xl|l-2xl|xl-3xl|2xl-4xl|3xl-5xl|4xl-6xl|5xl-6xl|6xl-7xl&g=s,l,xl,12

            'xs': 'clamp(0.6944rem, 0.6299rem + 0.3227vw, 0.88rem)',
            'sm': 'clamp(0.9375rem, 0.9159rem + 0.1085vw, 1rem)',
            'm': 'clamp(1rem, 0.8696rem + 0.6522vw, 1.375rem)',
            'lg': 'clamp(1.35rem, 1.2767rem + 0.3688vw, 1.5625rem)',
            'xl': 'clamp(1.62rem, 1.5051rem + 0.5781vw, 1.9531rem)',
            '2xl': 'clamp(1.9438rem, 1.7722rem + 0.8633vw, 2.4413rem)',
            '3xl': 'clamp(2.3325rem, 2.0844rem + 1.2484vw, 3.0519rem)',
            '4xl': 'clamp(2.7994rem, 2.4491rem + 1.7625vw, 3.815rem)',
        },

        spacing: {
            // @link https://utopia.fyi/space/calculator?c=320,16,1.2,1240,22,1.25,5,2,&s=0.75|0.5|0.25,1.5|2|3|4|6|8|11|15|19,3xs-xs|2xs-s|xs-m|s-l|m-xl|l-2xl|xl-3xl|2xl-4xl|3xl-5xl|4xl-6xl|5xl-6xl|6xl-7xl|s-2xl&g=s,l,xl,12

            '4xs': 'clamp(0.125rem, 0.1033rem + 0.1087vw, 0.1875rem)',
            '3xs': 'clamp(0.25rem, 0.2065rem + 0.2174vw, 0.375rem)',
            '2xs': 'clamp(0.5rem, 0.4348rem + 0.3261vw, 0.6875rem)',
            'xs': 'clamp(0.75rem, 0.6413rem + 0.5435vw, 1.0625rem)',
            's': 'clamp(1rem, 0.8696rem + 0.6522vw, 1.375rem)',
            'm': 'clamp(1.5rem, 1.3043rem + 0.9783vw, 2.0625rem)',
            'l': 'clamp(2rem, 1.7391rem + 1.3043vw, 2.75rem)',
            'xl': 'clamp(3rem, 2.6087rem + 1.9565vw, 4.125rem)',
            '2xl': 'clamp(4rem, 3.4783rem + 2.6087vw, 5.5rem)',
            '3xl': 'clamp(6rem, 5.2174rem + 3.913vw, 8.25rem)',
            '4xl': 'clamp(8rem, 6.9565rem + 5.2174vw, 11rem)',
            '5xl': 'clamp(11rem, 9.5652rem + 7.1739vw, 15.125rem)',
            '6xl': 'clamp(15rem, 13.0435rem + 9.7826vw, 20.625rem)',
            '7xl': 'clamp(19rem, 16.5217rem + 12.3913vw, 26.125rem)',

            // One-up pairs
            '3xs-2xs': 'clamp(0.25rem, 0.0978rem + 0.7609vw, 0.6875rem)',
            '2xs-xs': 'clamp(0.5rem, 0.3043rem + 0.9783vw, 1.0625rem)',
            'xs-s': 'clamp(0.75rem, 0.5326rem + 1.087vw, 1.375rem)',
            's-m': 'clamp(1rem, 0.6304rem + 1.8478vw, 2.0625rem)',
            'm-l': 'clamp(1.5rem, 1.0652rem + 2.1739vw, 2.75rem)',
            'l-xl': 'clamp(2rem, 1.2609rem + 3.6957vw, 4.125rem)',
            'xl-2xl': 'clamp(3rem, 2.1304rem + 4.3478vw, 5.5rem)',
            '2xl-3xl': 'clamp(4rem, 2.5217rem + 7.3913vw, 8.25rem)',
            '3xl-4xl': 'clamp(6rem, 4.2609rem + 8.6957vw, 11rem)',
            '4xl-5xl': 'clamp(8rem, 5.5217rem + 12.3913vw, 15.125rem)',
            '5xl-6xl': 'clamp(11rem, 7.6522rem + 16.7391vw, 20.625rem)',
            '6xl-7xl': 'clamp(15rem, 11.1304rem + 19.3478vw, 26.125rem)',

            // Two-up pairs
            '3xs-xs': 'clamp(0.25rem, -0.0326rem + 1.413vw, 1.0625rem)',
            '2xs-s': 'clamp(0.5rem, 0.1957rem + 1.5217vw, 1.375rem)',
            'xs-m': 'clamp(0.75rem, 0.2935rem + 2.2826vw, 2.0625rem)',
            's-l': 'clamp(1rem, 0.3913rem + 3.0435vw, 2.75rem)',
            'm-xl': 'clamp(1.5rem, 0.587rem + 4.5652vw, 4.125rem)',
            'l-2xl': 'clamp(2rem, 0.7826rem + 6.087vw, 5.5rem)',
            'xl-3xl': 'clamp(3rem, 1.1739rem + 9.1304vw, 8.25rem)',
            '2xl-4xl': 'clamp(4rem, 1.5652rem + 12.1739vw, 11rem)',
            '3xl-5xl': 'clamp(6rem, 2.8261rem + 15.8696vw, 15.125rem)',
            '4xl-6xl': 'clamp(8rem, 3.6087rem + 21.9565vw, 20.625rem)',
            '5xl-6xl': 'clamp(11rem, 7.6522rem + 16.7391vw, 20.625rem)',
            '6xl-7xl': 'clamp(15rem, 11.1304rem + 19.3478vw, 26.125rem)',

            // Custom
            's-2xl': 'clamp(1rem, -0.5652rem + 7.8261vw, 5.5rem)',
        },

        extend: {
            maxWidth: {
                'content': 'min(clamp(320px, 100%, max(680px, 33svw)), 100svw)',
            },

            keyframes: {
                underlineWave: {
                    '0%, 100%': {
                        'text-decoration-color': 'inherit',
                        'color': 'inherit',
                    },
                    '50%': {
                        'text-decoration-color': 'theme("colors.accent")',
                        'color': 'theme("colors.accent")',
                    },
                },
            },
        },
    },
    corePlugins: {
        preflight: true,
    },
}
