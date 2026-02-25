// Vérifier le statut de connexion de l'utilisateur
async function checkUserStatus() {
    try {
        const response = await fetch('/api/user');
        const data = await response.json();
        
        const authGuest = document.getElementById('auth-guest');
        const authUser = document.getElementById('auth-user');
        const userPrenomSpan = document.getElementById('user-prenom');
        
        if (!authGuest || !authUser || !userPrenomSpan) {
            console.error('Éléments auth manquants');
            return;
        }
        
        if (data.loggedIn) {
            authGuest.style.display = 'none';
            authUser.style.display = 'flex';
            userPrenomSpan.textContent = data.prenom || data.email;
        } else {
            authGuest.style.display = 'flex';
            authUser.style.display = 'none';
        }
    } catch (error) {
        console.error('Erreur vérification statut:', error);
    }
}

// Lancer la vérification au chargement de la page
document.addEventListener('DOMContentLoaded', checkUserStatus);
