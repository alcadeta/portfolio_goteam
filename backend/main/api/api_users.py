from rest_framework.decorators import api_view
from rest_framework.response import Response
from rest_framework.exceptions import ErrorDetail
from ..models import User, Board
from ..util import validate_team_id, validate_board_id, authenticate


@api_view(['GET', 'POST'])
def users(request):
    auth_user = request.META.get('HTTP_AUTH_USER')
    auth_token = request.META.get('HTTP_AUTH_TOKEN')

    authentication_response = authenticate(auth_user, auth_token)
    if authentication_response:
        return authentication_response

    if request.method == 'GET':
        team_id = request.query_params.get('team_id')
        validation_response = validate_team_id(team_id)
        if validation_response:
            return validation_response

        board_id = request.query_params.get('board_id')
        board, validation_response = validate_board_id(board_id)
        if validation_response:
            return validation_response

        users_list = User.objects.filter(team_id=team_id)
        board_users = User.objects.filter(board=board)

        return Response(list(map(
            lambda user: {'username': user.username,
                          'isActive': user in board_users},
            users_list
        )), 200)

    if request.method == 'POST':
        username = request.data.get('username')
        if not username:
            return Response({
                'username': ErrorDetail(string='Username cannot be empty.',
                                        code='blank')
            }, 400)

        board_id = request.data.get('board_id')
        is_active = request.data.get('is_active')

        user = User.objects.get(username=username)
        board = Board.objects.get(id=board_id)

        if is_active == 'True':
            board.user.add(user)
        else:
            board.user.remove(user)

        return Response({'msg': f'{username} is removed from {board.name}.'},
                        200)