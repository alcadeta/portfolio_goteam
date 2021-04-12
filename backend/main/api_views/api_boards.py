from rest_framework.decorators import api_view
from rest_framework.response import Response
from rest_framework.exceptions import ErrorDetail
from ..serializers.ser_board import BoardSerializer
from ..serializers.ser_column import ColumnSerializer
from ..models import Board, Team, User
import bcrypt


@api_view(['POST', 'GET'])
def boards(request):
    forbidden_response = Response({
        'auth': ErrorDetail(string="Authentication failure.",
                            code='not_authenticated')
    }, 403)

    # validate username
    username = request.META.get('HTTP_AUTH_USER')
    if not username:
        return forbidden_response
    try:
        user = User.objects.get(username=username)
    except User.DoesNotExist:
        return forbidden_response

    # validate authentication token
    token = request.META.get('HTTP_AUTH_TOKEN')
    if not token:
        return forbidden_response
    try:
        tokens_match = bcrypt.checkpw(
            bytes(user.username, 'utf-8') + user.password,
            bytes(token, 'utf-8'))
        if not tokens_match:
            return forbidden_response
    except ValueError:
        return forbidden_response

    if request.method == 'POST':
        # validate is_admin
        if not user.is_admin:
            return Response({
                'username': ErrorDetail(
                    string='Only the team admin can create a board.',
                    code='not_authorized'
                )
            }, 400)

        team_id = request.data.get('team_id')
        if not team_id:
            return Response({
                'team_id': ErrorDetail(string='Team ID cannot be empty.',
                                       code='blank')
            }, 400)
        try:
            Team.objects.get(id=team_id)
        except Team.DoesNotExist:
            return Response({
                'team_id': ErrorDetail(string='Team not found.',
                                       code='not_found')
            }, 404)

        # create board
        board_serializer = BoardSerializer(data={'team': team_id})
        if not board_serializer.is_valid():
            return Response(board_serializer.errors, 400)
        board = board_serializer.save()

        # create four columns for the board
        for order in range(0, 4):
            column_serializer = ColumnSerializer(
                data={'board': board.id, 'order': order}
            )
            if not column_serializer.is_valid():
                return Response(column_serializer.errors, 400)
            column_serializer.save()

        # return success response
        return Response({
            'msg': 'Board creation successful.',
            'board_id': board.id
        }, 201)

    if request.method == 'GET':
        # validate team_id
        team_id = request.query_params.get('team_id')
        if not team_id:
            return Response({
                'team_id': ErrorDetail(string='Team ID cannot be empty.',
                                       code='blank')
            }, 400)
        try:
            Team.objects.get(id=team_id)
        except Team.DoesNotExist:
            return Response({
                'team_id': ErrorDetail(string='Team not found.',
                                       code='not_found')
            }, 404)

        # create a board if none exists for the team
        team_boards = Board.objects.filter(team=team_id)
        if not team_boards:
            if user.is_admin:
                # create a board
                serializer = BoardSerializer(data={'team': team_id})
                if not serializer.is_valid():
                    return Response({
                        'team_id': ErrorDetail(string='Invalid team ID.',
                                               code='invalid')
                    }, 400)
                board = serializer.save()

                # return a list containing only the new board
                return Response({
                    'boards': [
                        {'board_id': board.id, 'team_id': board.team.id}
                    ]
                }, 201)

            return Response({
                'team_id': ErrorDetail(string='Boards not found.',
                                       code='not_found')
            }, 404)

        # return pre-existing boards
        serializer = BoardSerializer(team_boards, many=True)
        return Response({'boards': serializer.data}, 200)
