import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params }) => {
    // Tutaj w przyszłości dodasz pobieranie danych o pokoju z backendu po slugu
    // const res = await fetch(`/api/v1/room/${params.slug}`);
    
    return {
        slug: params.slug
    };
};
