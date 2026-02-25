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
        button.addEventListener('click', async function(e) {
            e.preventDefault();
            console.log('Bouton devis cliqué');
            
            const productName = this.getAttribute('data-product');
            
            // Vérifier si l'utilisateur est connecté
            try {
                const userResponse = await fetch('/api/user');
                const userData = await userResponse.json();
                
                if (!userData.loggedIn) {
                    // Afficher le message de connexion requise dans la modal
                    showLoginRequired(productName);
                    return;
                }
            } catch (error) {
                console.error('Erreur vérification connexion:', error);
            }
            
            // Utilisateur connecté - afficher le formulaire
            document.getElementById('productName').value = productName;
            document.getElementById('modalProductTitle').textContent = productName;
            document.querySelector('.quote-modal-body').style.display = 'block';
            
            // Cacher le message de connexion s'il existe
            const loginMsg = document.getElementById('loginRequiredMessage');
            if (loginMsg) {
                loginMsg.style.display = 'none';
            }
            
            modal.style.display = 'flex';
        });
    });

    // Afficher le message de connexion requise
    function showLoginRequired(productName) {
        document.getElementById('modalProductTitle').textContent = productName;
        document.querySelector('.quote-modal-body').style.display = 'none';
        
        let loginMsg = document.getElementById('loginRequiredMessage');
        if (!loginMsg) {
            loginMsg = document.createElement('div');
            loginMsg.id = 'loginRequiredMessage';
            loginMsg.innerHTML = `
                <div style="text-align: center; padding: 2rem;">
                    <p style="font-size: 1.1rem; margin-bottom: 1.5rem; color: #333;">
                        Vous devez être connecté pour demander un devis
                    </p>
                    <div style="display: flex; gap: 1rem; justify-content: center;">
                        <a href="/login.html" class="quote-form-submit" style="display: inline-block; text-decoration: none;">Se connecter</a>
                        <a href="/register.html" class="quote-form-submit" style="display: inline-block; text-decoration: none; background: #6161AB;">S'inscrire</a>
                    </div>
                </div>
            `;
            document.querySelector('.quote-modal-content').appendChild(loginMsg);
        }
        
        loginMsg.style.display = 'block';
        modal.style.display = 'flex';
    }

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
            // 1. Enregistrer dans la base de données
            const dbResponse = await fetch('/api/quote', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    nom: nom,
                    prenom: prenom,
                    email: email,
                    telephone: telephone,
                    produit: produit,
                    message: message
                })
            });

            // Vérifier si l'utilisateur n'est pas connecté
            if (dbResponse.status === 401) {
                const data = await dbResponse.json();
                alert(data.message || 'Vous devez être connecté pour demander un devis. Veuillez vous connecter d\'abord.');
                modal.style.display = 'none';
                form.reset();
                submitBtn.disabled = false;
                submitBtn.textContent = originalText;
                return;
            }

            if (!dbResponse.ok) {
                throw new Error('Erreur lors de l\'enregistrement');
            }

            // 2. Envoyer via FormSubmit.co
            await fetch('https://formsubmit.co/elsachochon13@gmail.com', {
                method: 'POST',
                body: formData,
                mode: 'no-cors'
            });

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
