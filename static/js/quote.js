// Gestion de la modal de demande de devis
document.addEventListener('DOMContentLoaded', function() {
    const modal = document.getElementById('quoteModal');
    const closeBtn = document.querySelector('.quote-modal-close');
    const form = document.getElementById('quoteForm');
    const openButtons = document.querySelectorAll('.open-quote-modal');

    // Vérifier que tous les éléments existent
    if (!modal || !closeBtn || !form || openButtons.length === 0) {
        console.error('Éléments manquants pour la modal de devis');
        return;
    }

    console.log('Modal de devis initialisée avec', openButtons.length, 'bouton(s)');

    // Ouvrir la modal
    openButtons.forEach(button => {
        button.addEventListener('click', function(e) {
            e.preventDefault();
            console.log('Bouton devis cliqué');
            const productName = this.getAttribute('data-product');
            document.getElementById('productName').value = productName;
            document.getElementById('modalProductTitle').textContent = productName;
            modal.style.display = 'flex';
        });
    });

    // Fermer la modal
    closeBtn.addEventListener('click', function() {
        modal.style.display = 'none';
        form.reset();
    });

    // Fermer en cliquant à l'extérieur
    window.addEventListener('click', function(e) {
        if (e.target === modal) {
            modal.style.display = 'none';
            form.reset();
        }
    });

    // Soumettre le formulaire
    form.addEventListener('submit', async function(e) {
        e.preventDefault();

        const submitBtn = form.querySelector('button[type="submit"]');
        const originalText = submitBtn.textContent;
        submitBtn.disabled = true;
        submitBtn.textContent = 'Envoi en cours...';

        const nom = document.getElementById('nom').value;
        const prenom = document.getElementById('prenom').value;
        const email = document.getElementById('email').value;
        const telephone = document.getElementById('telephone').value;
        const produit = document.getElementById('productName').value;

        // Créer le message
        const message = `Bonjour,

J'aimerais demander un devis pour le produit : ${produit}

Mes coordonnées :
- Nom : ${nom}
- Prénom : ${prenom}
- Email : ${email}
- Téléphone : ${telephone}

Merci de me renvoyer le devis pour ce produit.

Cordialement,
${prenom} ${nom}`;

        // Utiliser FormSubmit.co pour envoyer l'email directement
        const formData = new FormData();
        formData.append('_subject', `Demande de devis - ${produit}`);
        formData.append('_email', email);
        formData.append('_template', 'table');
        formData.append('_captcha', 'false');
        formData.append('message', message);
        formData.append('nom', nom);
        formData.append('prenom', prenom);
        formData.append('telephone', telephone);
        formData.append('produit', produit);

        try {
            // Envoyer via FormSubmit.co
            const response = await fetch('https://formsubmit.co/elsachochon13@gmail.com', {
                method: 'POST',
                body: formData,
                mode: 'no-cors'
            });

            // no-cors ne permet pas de lire la réponse, on suppose que ça marche
            alert('Votre demande de devis a été envoyée avec succès ! Nous vous recontacterons rapidement.');
            modal.style.display = 'none';
            form.reset();
        } catch (error) {
            console.error('Erreur:', error);
            alert('Une erreur est survenue. Veuillez réessayer.');
        } finally {
            submitBtn.disabled = false;
            submitBtn.textContent = originalText;
        }
    });
});
