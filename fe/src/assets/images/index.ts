// Photos from Unsplash (Unsplash License). Downloads were registered through the
// Unsplash API download endpoint; credit is shown next to each photo.
import type { StaticImageData } from "next/image";
import heroResilientImg from "./hero-resilient.webp";
import featureLandscapeImg from "./feature-landscape.webp";
import loginDeskImg from "./login-desk.webp";

export type TSitePhoto = {
  src: StaticImageData;
  alt: string;
  author: string;
  authorUrl: string;
  photoUrl: string;
};

export const sitePhotos = {
  heroResilient: {
    src: heroResilientImg,
    alt: "A small plant growing out of a crack in a rock",
    author: "Fabian Kleiser",
    authorUrl: "https://unsplash.com/@fabiankleiser",
    photoUrl: "https://unsplash.com/photos/a-small-plant-growing-out-of-a-crack-in-a-rock-CIdlJd0ZbrU",
  },
  featureLandscape: {
    src: featureLandscapeImg,
    alt: "The sun rising over a mountain lake surrounded by pine trees",
    author: "Colin Lloyd",
    authorUrl: "https://unsplash.com/@onthesearchforpineapples",
    photoUrl: "https://unsplash.com/photos/the-sun-shines-on-a-mountain-lake-surrounded-by-pine-trees-4x8EBg02E6I",
  },
  loginDesk: {
    src: loginDeskImg,
    alt: "An open notebook with a pen on a wooden desk",
    author: "Gabriel Cox",
    authorUrl: "https://unsplash.com/@pendulumphotographs",
    photoUrl: "https://unsplash.com/photos/an-open-notebook-with-a-pen-on-a-wooden-desk-6e7F96dBlAA",
  },
} satisfies Record<string, TSitePhoto>;
